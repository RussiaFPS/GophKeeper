package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	delivery "gopthkeeper/internal/delivery/client"
	"gopthkeeper/internal/models"
	"net/http"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Store a new secret",
	Long: `Store a new secret of a specified type (login/password, text, binary, bank card)
on the GophKeeper server. Requires authentication.`,
	Run: func(cmd *cobra.Command, args []string) {
		secretTypeStr, _ := cmd.Flags().GetString("type")
		dataStr, _ := cmd.Flags().GetString("data")
		metadata, _ := cmd.Flags().GetString("metadata")
		secretID, _ := cmd.Flags().GetInt("id") // 0 if not provided

		if secretTypeStr == "" || dataStr == "" {
			fmt.Println("Error: Secret type and data cannot be empty.")
			cmd.Help()
			return
		}

		var secretType models.SecretType
		switch secretTypeStr {
		case "login":
			secretType = models.LoginPasswordType
		case "text":
			secretType = models.TextDataType
		case "binary":
			secretType = models.BinaryDataType
		case "bankcard":
			secretType = models.BankCardType
		default:
			fmt.Printf("Error: Invalid secret type '%s'. Valid types are: login, text, binary, bankcard.\n", secretTypeStr)
			return
		}

		secret := models.Secret{
			Type:     secretType,
			Data:     []byte(dataStr),
			Metadata: metadata,
		}

		client := delivery.NewClient()
		var resp *http.Response
		var err error

		if secretID != 0 {
			// Update existing secret
			secret.ID = secretID
			resp, err = client.AuthenticatedRequest(http.MethodPut, fmt.Sprintf("/api/secrets/%d", secretID), secret)
		} else {
			// Create new secret
			resp, err = client.AuthenticatedRequest(http.MethodPost, "/api/secrets", secret)
		}

		if err != nil {
			fmt.Printf("Error sending request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			fmt.Printf("Operation failed: %s (Status: %d)\n", buf.String(), resp.StatusCode)
			return
		}

		var resultSecret models.Secret
		if err := json.NewDecoder(resp.Body).Decode(&resultSecret); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			return
		}

		if secretID != 0 {
			fmt.Printf("Secret ID %d updated successfully!\n", resultSecret.ID)
		} else {
			fmt.Printf("Secret created successfully with ID: %d\n", resultSecret.ID)
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().StringP("type", "t", "", "Type of secret (login, text, binary, bankcard)")
	setCmd.Flags().StringP("data", "d", "", "The secret data to store")
	setCmd.Flags().StringP("metadata", "m", "", "Optional metadata for the secret")
	setCmd.Flags().IntP("id", "i", 0, "Optional: ID of the secret to update (if omitted, creates a new secret)")

	setCmd.MarkFlagRequired("type")
	setCmd.MarkFlagRequired("data")
}
