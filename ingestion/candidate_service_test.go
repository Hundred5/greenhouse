package ingestion

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_candidateService_Retrieve(t *testing.T) {
	setup()
	defer teardown()

	type args struct {
		ids []int64
	}
	test := struct {
		args           args
		wantCandidates []Candidate
		wantErr        bool
	}{
		args: args{
			ids: []int64{12},
		},
		wantCandidates: []Candidate{
			Candidate{
				ID:         17681532,
				Name:       "Harry Potter",
				ExternalID: "24680",
				Applications: []Application{
					Application{
						ID:         59724,
						Job:        "Auror",
						Status:     "Active",
						Stage:      "Application Review",
						ProfileURL: "https://app.greenhouse.io/people/17681532?application_id=26234709",
					},
				},
			},
		},
	}

	mux.HandleFunc("/v1/partner/candidates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.WriteHeader(200)
		io.WriteString(w, `
		[
			{
				"id": 17681532,
				"name": "Harry Potter",
				"external_id": "24680",
				"applications": [
					{
						"id": 59724,
						"job": "Auror",
						"status": "Active",
						"stage": "Application Review",
						"profile_url": "https://app.greenhouse.io/people/17681532?application_id=26234709"
					}
				]
			}
		]
		`)
	})

	gotCandidates, err := client.Candidates.Retrieve(test.args.ids)

	switch test.wantErr {
	case true:
		assert.Error(t, err)
	case false:
		assert.NoError(t, err)
	}
	assert.Equal(t, test.wantCandidates, gotCandidates)

}

func Test_candidateService_Post(t *testing.T) {
	setup()
	defer teardown()

	type args struct {
		candidates []PostCandidate
	}
	test := struct {
		args           args
		reqBody        string
		wantCandidates []PostCandidateResponse
		wantErr        bool
	}{
		reqBody: `[{"prospect":true,"first_name":"Harry","last_name":"Potter","company":"Hogwarts","title":"Student","resume":"https://hogwarts.com/resume","phone_numbers":[{"phone_number":"123-456-7890","type":"home"}],"emails":[{"email":"hpotter@hogwarts.edu","type":"other"}],"social_media":[{"url":"https://twitter.com/hp"}],"websites":[{"url":"https://harrypotter.com","type":"blog"}],"addresses":[{"address":"4 Privet Dr","type":"home"}],"job_id":12345,"external_id":"24680","notes":"Good at Quiddich","prospect_pool_id":123,"prospect_pool_stage_id":456,"prospect_owner_email":"prospect_owners_email@example.com"}]`,
		args: args{
			candidates: []PostCandidate{
				PostCandidate{
					Prospect:  true,
					FirstName: "Harry",
					LastName:  "Potter",
					Title:     "Student",
					Company:   "Hogwarts",
					Resume:    "https://hogwarts.com/resume",
					PhoneNumbers: []PhoneNumber{
						PhoneNumber{
							PhoneNumber: "123-456-7890",
							Type:        PhoneNumberTypeHome,
						},
					},
					Emails: []Email{
						Email{
							Email: "hpotter@hogwarts.edu",
							Type:  EmailTypeOther,
						},
					},
					SocialMedia: []SocialMedia{
						SocialMedia{
							URL: "https://twitter.com/hp",
						},
					},
					Websites: []Website{
						Website{
							URL:  "https://harrypotter.com",
							Type: WebsiteTypeBlog,
						},
					},
					Addresses: []Address{
						Address{
							Address: "4 Privet Dr",
							Type:    AddressTypeHome,
						},
					},
					JobID:               12345,
					ExternalID:          "24680",
					Notes:               "Good at Quiddich",
					ProspectPoolID:      123,
					ProspectPoolStageID: 456,
					ProspectOwnerEmail:  "prospect_owners_email@example.com",
				},
			},
		},
		wantCandidates: []PostCandidateResponse{
			PostCandidateResponse{
				ID:            12345,
				ApplicationID: 17681532,
				ExternalID:    "24680",
				ProfileURL:    "https://app.greenhouse.io/people/17681532?application_id=26234709",
			},
		},
	}

	mux.HandleFunc("/v1/partner/candidates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		body := formatReadCloser(&r.Body)
		equal, err := areEqualJSON(test.reqBody, body)
		assert.NoError(t, err)
		if !equal {
			assert.Equal(t, test.reqBody, body) //just to get the diff
		}
		w.WriteHeader(200)
		io.WriteString(w, `
		[
			{
				"id": 12345,
				"application_id": 17681532,
				"external_id": "24680",
				"profile_url": "https://app.greenhouse.io/people/17681532?application_id=26234709"
			}
		]
		`)
	})

	gotCandidates, err := client.Candidates.Post(test.args.candidates)

	switch test.wantErr {
	case true:
		assert.Error(t, err)
	case false:
		assert.NoError(t, err)
	}
	assert.Equal(t, test.wantCandidates, gotCandidates)

}
