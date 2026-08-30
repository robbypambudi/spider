package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spider/spider/pkg/config"
)

func main() {
	settings, _ := config.Load()
	base := settings.APIBaseURL

	var token string
	root := &cobra.Command{Use: "spider"}
	root.PersistentFlags().StringVar(&token, "token", "", "Bearer token (user auth, from /auth/login)")

	security := &cobra.Command{Use: "security"}
	scan := &cobra.Command{
		Use:  "scan [text]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := args[0]
			return apiPost(base+"/api/v1/security/scan", token, map[string]string{"text": text}, os.Stdout)
		},
	}
	security.AddCommand(scan)
	root.AddCommand(security)
	root.AddCommand(workerCommand(settings, base, &token))

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func apiPost(url, token string, body interface{}, out *os.File) error {
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var pretty json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&pretty); err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(pretty, "", "  ")
	_, err = fmt.Fprintln(out, string(enc))
	return err
}

func apiGet(url, token string, out *os.File) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var pretty json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&pretty); err != nil {
		return err
	}
	enc, _ := json.MarshalIndent(pretty, "", "  ")
	_, err = fmt.Fprintln(out, string(enc))
	return err
}
