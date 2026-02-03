package services

import (
	"encoding/json"
	"fmt"
	delivery "gopthkeeper/internal/delivery/client"
	"net/http"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a secret",
	Long:  `Delete a specific secret by its ID from the GophKeeper server. Requires authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		secretID, _ := cmd.Flags().GetInt("id")

		if secretID == 0 {
			fmt.Println("Error: Secret ID is required for deletion.")
			cmd.Help()
			return
		}

		client := delivery.NewClient()
		resp, err := client.AuthenticatedRequest(http.MethodDelete, fmt.Sprintf("/api/secrets/%d", secretID), nil)
		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			// Read the response body for more detailed error message
			// (even for 404, the server handler might write a message)
			// Assuming the server returns JSON error messages
			var errMessage map[string]string
			if json.NewDecoder(resp.Body).Decode(&errMessage) == nil {
				if msg, ok := errMessage["message"]; ok {
					fmt.Printf("Deletion failed: %s (Status: %d)\n", msg, resp.StatusCode)
					return
				}
			}
			fmt.Printf("Deletion failed: Unexpected status code %d\n", resp.StatusCode)
			return
		}

		fmt.Printf("Secret ID %d deleted successfully!\n", secretID)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().IntP("id", "i", 0, "ID of the secret to delete")
	deleteCmd.MarkFlagRequired("id")
}
