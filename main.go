package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	showLabels = false
	user       = "me"
	readLabel  = "Label_3754174685129828981"
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
			batchMsgID = append(batchMsgID, m.Id)
			if resumeType == 1 {
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
			AddLabelIds:    []string{readLabel},
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
