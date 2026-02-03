package services

import (
	"encoding/json"
	"fmt"
	delivery "gopthkeeper/internal/delivery/client"
	"gopthkeeper/internal/models"
	"io"
	"net/http"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve secrets",
	Long: `Retrieve all secrets or a specific secret by ID from the GophKeeper server.
Requires authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		secretID, _ := cmd.Flags().GetInt("id")

		client := delivery.NewClient()
		var resp *http.Response
		var err error

		if secretID != 0 {
			// Get specific secret by ID
			resp, err = client.AuthenticatedRequest(http.MethodGet, fmt.Sprintf("/api/secrets/%d", secretID), nil)
		} else {
			// Get all secrets
			resp, err = client.AuthenticatedRequest(http.MethodGet, "/api/secrets", nil)
		}

		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("Operation failed: %s (Status: %d)\n", string(bodyBytes), resp.StatusCode)
			return
		}

		if secretID != 0 {
			var secret models.Secret
			if err := json.NewDecoder(resp.Body).Decode(&secret); err != nil {
				fmt.Printf("Error decoding secret: %v\n", err)
				return
			}
			fmt.Printf("Secret ID: %d, Type: %s, Data: %s, Metadata: %s\n", secret.ID, secret.Type.String(), string(secret.Data), secret.Metadata)
		} else {
			var secrets []models.Secret
			if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
				fmt.Printf("Error decoding secrets: %v\n", err)
				return
			}
			if len(secrets) == 0 {
				fmt.Println("No secrets found.")
				return
			}
			fmt.Println("Your secrets:")
			for _, secret := range secrets {
				fmt.Printf("  ID: %d, Type: %s, Data: %s, Metadata: %s\n", secret.ID, secret.Type.String(), string(secret.Data), secret.Metadata)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().IntP("id", "i", 0, "Optional: ID of the secret to retrieve")
}
