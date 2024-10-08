package handlers

import (
	"api/pkg/database"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Parse the request body as JSON
	var requestData map[string]interface{}
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Error parsing request data", http.StatusBadRequest)
		return
	}

	// Extract email, password, firstName, and lastName from the request data
	email, emailExists := requestData["email"].(string)
	password, passwordExists := requestData["password"].(string)
	firstName, firstNameExists := requestData["firstName"].(string)
	lastName, lastNameExists := requestData["lastName"].(string)

	// Check if the required fields exist
	if !emailExists || !passwordExists || !firstNameExists || !lastNameExists {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Construct the JSON payload for Auth0 API
	data := map[string]interface{}{
		"email":          email,
		"user_metadata":  map[string]interface{}{},
		"blocked":        false,
		"email_verified": false, // is the email verified?
		"app_metadata":   map[string]interface{}{},
		"given_name":     firstName,
		"family_name":    lastName,
		"name":           firstName + " " + lastName,
		"connection":     "Username-Password-Authentication",
		"password":       password,
		"verify_email":   true, // Send a verification email to the user
	}

	formDataBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError)
		return
	}

	token, err := GenerateAuth0Token()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		http.Error(w, "Error getting token", http.StatusInternalServerError)
		return
	}

	// Create a POST request to the users endpoint
	req, err := http.NewRequest("POST", "https://"+os.Getenv("AUTH0_DOMAIN")+"/api/v2/users", bytes.NewBuffer(formDataBytes))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("authorization", "Bearer "+token)

	// Send the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		http.Error(w, string(bodyBytes), resp.StatusCode)
		return
	}

	// Write the successful response to the client
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	// Parse the response body
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		http.Error(w, "Error parsing response data", http.StatusInternalServerError)
		return
	}
	userId, ok := result["user_id"].(string)
	if !ok {
		http.Error(w, "Error retrieving user_id", http.StatusInternalServerError)
		return
	}

	// Add the user to the database
	if err := database.CreateUser(userId, email, firstName, lastName); err != nil {
		http.Error(w, "Error adding user to database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(bodyBytes)
}
