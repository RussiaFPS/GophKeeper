package services

import (
	"bytes"
	"fmt"
	delivery "gopthkeeper/internal/delivery/client"
	"gopthkeeper/internal/models"
	"net/http"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long:  `Register a new user with a username and password on the GophKeeper server.`,
	Run: func(cmd *cobra.Command, args []string) {
		login, _ := cmd.Flags().GetString("login")
		password, _ := cmd.Flags().GetString("password")

		if login == "" || password == "" {
			fmt.Println("Error: Login and password cannot be empty.")
			cmd.Help()
			return
		}

		user := models.User{
			Login:    login,
			Password: password,
		}

		client := delivery.NewClient()
		resp, err := client.Request(http.MethodPost, "/api/user/register", user)
		if err != nil {
			fmt.Printf("Error sending registration request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			fmt.Printf("Registration failed: %s\n", buf.String())
			return
		}

		fmt.Println("User registered successfully!")
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)

	registerCmd.Flags().StringP("login", "l", "", "User login/username")
	registerCmd.Flags().StringP("password", "p", "", "User password")
	registerCmd.MarkFlagRequired("login")
	registerCmd.MarkFlagRequired("password")
}
