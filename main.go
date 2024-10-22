package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
)

const (
    tenantID     = "your-tenant-id"
    clientID     = "your-client-id"
    clientSecret = "your-client-secret"
    appID        = "cf1256c5-fe39-4511-b023-47daa5df8946" // The ID of the app you want to add as a tab
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

func addTabToChannel(accessToken, teamID, channelID string) error {
    url := fmt.Sprintf("https://graph.microsoft.com/v1.0/teams/%s/channels/%s/tabs", teamID, channelID)
    tab := map[string]interface{}{
        "displayName": "My Tab",
        "teamsApp@odata.bind": fmt.Sprintf("https://graph.microsoft.com/v1.0/appCatalogs/teamsApps/%s", appID),
        "configuration": map[string]string{
            "entityId":    "your-entity-id",
            "contentUrl":  "https://yourwebsite.com/your-tab-content",
            "websiteUrl":  "https://yourwebsite.com",
            "removeUrl":   "https://yourwebsite.com/remove",
        },
    }
    tabData, err := json.Marshal(tab)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(tabData))
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
        return fmt.Errorf("failed to add tab: %s", resp.Status)
    }
    return nil
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Unable to read request body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    var notification map[string]interface{}
    if err := json.Unmarshal(body, &notification); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    fmt.Printf("Received notification: %v\n", notification)
    // Extract teamID and channelID from the notification
    teamID := notification["teamId"].(string)
    channelID := notification["channelId"].(string)

    accessToken, err := getAccessToken()
    if err != nil {
        fmt.Println("Error getting access token:", err)
        return
    }

    err = addTabToChannel(accessToken, teamID, channelID)
    if err != nil {
        fmt.Println("Error adding tab to channel:", err)
        return
    }

    fmt.Println("Tab added successfully")
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    fmt.Printf("Webhook server started at :%s\n", port)
    http.ListenAndServe(":"+port, nil)
}