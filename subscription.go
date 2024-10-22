package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

const (
    tenantID     = "your-tenant-id"
    clientID     = "your-client-id"
    clientSecret = "your-client-secret"
    webhookURL   = "https://your-webhook-url/webhook"
)

func getAccessToken() (string, error) {
    url := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)
    data := "grant_type=client_credentials&client_id=" + clientID + "&client_secret=" + clientSecret + "&scope=https%3A%2F%2Fgraph.microsoft.com%2F.default"
    req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    return result["access_token"].(string), nil
}

func createSubscription(accessToken string) error {
    url := "https://graph.microsoft.com/v1.0/subscriptions"
    subscription := map[string]interface{}{
        "changeType": "created",
        "notificationUrl": webhookURL,
        "resource": "teams/getAllMessages",
        "expirationDateTime": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
        "clientState": "secretClientValue",
    }
    subscriptionData, err := json.Marshal(subscription)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(subscriptionData))
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+accessToken)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to create subscription: %s", resp.Status)
    }
    return nil
}

func main() {
    accessToken, err := getAccessToken()
    if err != nil {
        fmt.Println("Error getting access token:", err)
        os.Exit(1)
    }

    err = createSubscription(accessToken)
    if err != nil {
        fmt.Println("Error creating subscription:", err)
        os.Exit(1)
    }

    fmt.Println("Subscription created successfully")
}