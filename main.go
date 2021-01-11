package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

var (
	showLabels = false
	user       = "me"
)

func main() {
	flag.BoolVar(&showLabels, "show-labels", false, "Show Labels")
	flag.Parse()

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	if showLabels {
		lb, _ := srv.Users.Labels.List(user).Do()
		for _, l := range lb.Labels {
			fmt.Println(l.Id, l.Name)
		}
		return
	}

	conditions := "label:hiring/external-applications has:attachment is:unread"
	r, err := srv.Users.Messages.List(user).Q(conditions).Do()
	var firstRequest bool = true

	for {
		if err != nil {
			log.Fatalf("Unable to retrieve labels: %v", err)
		}
		if len(r.Messages) == 0 {
			if firstRequest {
				fmt.Println("No more messages found.")
				os.Exit(1)
			} else {
				fmt.Println("No messages found.")
				return
			}
		}
		firstRequest = false
		var batchMsgID []string
		for _, m := range r.Messages {
			msg, _ := srv.Users.Messages.Get(user, m.Id).Do()
			subject := getFromHeader(msg.Payload.Headers, "Subject")
			resumeType := validZhipinResumeMessage(subject)
			if resumeType == 1 {
				batchMsgID = append(batchMsgID, m.Id)
				var attID, attName string
				parts := msg.Payload.Parts
				for _, p := range parts {
					if len(p.Body.AttachmentId) > 0 {
						attID = p.Body.AttachmentId
						attName = p.Filename
					}
				}
				if len(attID) > 0 {
					file, _ := srv.Users.Messages.Attachments.Get(user, m.Id, attID).Do()
					resumeTypeStr := getResumeType(resumeType)
					err = downloadAttachments(msg.Id, resumeTypeStr, attName, file)
					if err == nil {
						fmt.Println("Got One Resume:", msg.Id, resumeTypeStr, attName)
					} else {
						fmt.Println(err)
					}
				} else {
					fmt.Println("Skip Message:", msg.Id)
				}
			} else {
				fmt.Println("Skipped non backend resumes:", msg.Id, resumeType)
			}
		}
		fmt.Println("Batch Read and Label for", batchMsgID)
		batchModifyReq := &gmail.BatchModifyMessagesRequest{
			Ids:            batchMsgID,
			AddLabelIds:    []string{"Label_3754174685129828981"},
			RemoveLabelIds: []string{"UNREAD"},
		}
		err = srv.Users.Messages.BatchModify(user, batchModifyReq).Do()
		if err != nil {
			fmt.Println(err)
		}

		nextPage := r.NextPageToken
		if nextPage == "" {
			break
		}
		fmt.Println("New Page: ", nextPage)
		r, err = srv.Users.Messages.List(user).PageToken(nextPage).Q(conditions).Do()
	}
}

func downloadAttachments(msgID string, resumeTypeStr string, fn string, att *gmail.MessagePartBody) error {
	var dst []byte
	dst, err := base64.URLEncoding.DecodeString(att.Data)
	if err != nil {
		fmt.Println(err)
		return err
	}
	realFn := fmt.Sprintf("./attachments/%s/%s_%s", resumeTypeStr, msgID, fn)
	err = ioutil.WriteFile(realFn, dst, 777)

	return err
}

func getResumeType(resumeType int) string {
	switch resumeType {
	case 1:
		return "backend"
	case 2:
		return "frontend"
	}
	return "unknown"
}

func validZhipinResumeMessage(subject string) int {
	if strings.HasPrefix(subject, "[External]") && strings.HasSuffix(subject, "【BOSS直聘】") {
		if strings.Contains(subject, "后端") {
			return 1
		} else if strings.Contains(subject, "前端") {
			return 2
		}
	}
	return 0
}

func getFromHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if h.Name == name {
			return h.Value
		}
	}
	return ""
}
