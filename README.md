# Gmail Resume Importer

Read the unread hiring mails and import attachments automatically from Gmail.

Not for common use. You may modify and use at your own risk.

## Prerequisite

1. Create New Label: `job-resumes`
2. Create New Label: `bot-downloaded` under `job-resumes`
3. Run and authorize with OAuth 2.0

## Usage

Show labels:

```sh
go run main.go -show-labels
```

Get resumes:

```sh
go run main.go
```
