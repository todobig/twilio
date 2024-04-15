package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	accountSid = "x" // replace with your Account SID
	authToken  = "x"   // replace with your Auth Token
	trunkSid   = "x" // replace with your Trunk SID
	addressSid = "x" // replace with your Address SID, required for regulatory compliance
)

type TwilioResponse struct {
	AvailablePhoneNumbers []struct {
		PhoneNumber string `json:"phone_number"`
	} `json:"available_phone_numbers"`
}

type PurchaseResponse struct {
	Sid string `json:"sid"`
}

func main() {
	urlValues := url.Values{}
	urlValues.Set("Country", "GB")
	urlValues.Set("Type", "mobile")
	urlValues.Set("PageSize", "10")

	req, _ := http.NewRequest("GET", "https://api.twilio.com/2010-04-01/Accounts/"+accountSid+"/AvailablePhoneNumbers/GB/Mobile.json?"+urlValues.Encode(), nil)
	req.SetBasicAuth(accountSid, authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	var twilioResponse TwilioResponse
	if err := json.NewDecoder(resp.Body).Decode(&twilioResponse); err != nil {
		fmt.Println(err)
		return
	}

	if len(twilioResponse.AvailablePhoneNumbers) >= 7 {
		for i := 0; i < 7; i++ {
			phoneNumber := twilioResponse.AvailablePhoneNumbers[i]

			data := url.Values{}
			data.Set("PhoneNumber", phoneNumber.PhoneNumber)
			data.Set("AddressSid", addressSid) // Include AddressSid in the POST data

			purchaseReq, _ := http.NewRequest("POST", "https://api.twilio.com/2010-04-01/Accounts/"+accountSid+"/IncomingPhoneNumbers.json", bytes.NewBufferString(data.Encode()))
			purchaseReq.SetBasicAuth(accountSid, authToken)
			purchaseReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			purchaseResp, err := client.Do(purchaseReq)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer purchaseResp.Body.Close()

			var purchaseResponse PurchaseResponse
			if err := json.NewDecoder(purchaseResp.Body).Decode(&purchaseResponse); err != nil {
				fmt.Println(err)
				return
			}

			// Assign the purchased phone number to a specific trunk
			updateData := url.Values{}
			updateData.Set("TrunkSid", trunkSid)
			updateReq, _ := http.NewRequest("POST", fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/IncomingPhoneNumbers/%s.json", accountSid, purchaseResponse.Sid), bytes.NewBufferString(updateData.Encode()))
			updateReq.SetBasicAuth(accountSid, authToken)
			updateReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			updateResp, err := client.Do(updateReq)
			if err != nil {
				fmt.Println("Error on updating phone number with trunk:\n[ERROR] -", err)
			} else {
				fmt.Printf("Purchased and assigned phone number %s with SID %s to trunk %s. Update status: %s\n", phoneNumber.PhoneNumber, purchaseResponse.Sid, trunkSid, updateResp.Status)
			}
		}
	} else {
		fmt.Println("Less than 7 available phone numbers.")
	}
}
