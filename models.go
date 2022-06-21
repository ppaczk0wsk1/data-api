package main

type DataWorldwide struct {
	Country          string `json:"country"`
	Date             string `json:"date"`
	Confirmed        string `json:"confirmed"`
	Deaths           string `json:"deaths"`
	DeathsPercent    string `json:"deaths_percent"`
	Recovered        string `json:"recovered"`
	RecoveredPercent string `json:"recovered_percent"`
	Active           string `json:"active"`
	ActivePercent    string `json:"active_percent"`
}

type DataCountries struct {
	Country          string `json:"country"`
	Date             string `json:"date"`
	Confirmed        string `json:"confirmed"`
	Deaths           string `json:"deaths"`
	DeathsPercent    string `json:"deaths_percent"`
	Recovered        string `json:"recovered"`
	RecoveredPercent string `json:"recovered_percent"`
	Active           string `json:"active"`
	ActivePercent    string `json:"active_percent"`
}

type DataCountriesDelta struct {
	Country               string `json:"country"`
	Date                  string `json:"date"`
	DeltaConfirmed        string `json:"delta_confirmed"`
	Confirmed             string `json:"confirmed"`
	DeltaConfirmedPercent string `json:"delta_confirmed_percent"`
	DeltaDeaths           string `json:"delta_deaths"`
	Deaths                string `json:"deaths"`
	DeltaDeathsPercent    string `json:"delta_deaths_percent"`
	DeltaRecovered        string `json:"delta_recovered"`
	Recovered             string `json:"recovered"`
	DeltaRecoveredPercent string `json:"delta_recovered_percent"`
	DeltaActive           string `json:"delta_active"`
	Active                string `json:"active"`
	DeltaActivePercent    string `json:"delta_active_percent"`
}

type DataConfirmed struct {
	Country          string `json:"country"`
	Date             string `json:"date"`
	Confirmed        string `json:"confirmed"`
	ConfirmedPercent string `json:"confirmed_percent"`
	TotalConfirmed   string `json:"total_confirmed"`
}

type DataDeltaConfirmed struct {
	Country               string `json:"country"`
	Date                  string `json:"date"`
	Confirmed             string `json:"confirmed"`
	DeltaConfirmed        string `json:"delta_confirmed"`
	DeltaConfirmedPercent string `json:"delta_confirmed_percent"`
	DeltaTotalConfirmed   string `json:"delta_total_confirmed"`
}

type DataDeltaActive struct {
	Country            string `json:"country"`
	Date               string `json:"date"`
	Confirmed          string `json:"confirmed"`
	Deaths             string `json:"deaths"`
	Recovered          string `json:"recovered"`
	Active             string `json:"active"`
	DeltaActive        string `json:"delta_active"`
	DeltaActivePercent string `json:"delta_active_percent"`
}

type DataCountriesList struct {
	Country string `json:"country"`
}

type DataCountriesDeltaPercent struct {
	Country               string `json:"country"`
	Date                  string `json:"date"`
	DeltaConfirmed        string `json:"delta_confirmed"`
	Confirmed             string `json:"confirmed"`
	DeltaConfirmedPercent string `json:"delta_confirmed_percent"`
	DeltaDeaths           string `json:"delta_deaths"`
	Deaths                string `json:"deaths"`
	DeathsPercent         string `json:"deaths_percent"`
	DeltaDeathsPercent    string `json:"delta_deaths_percent"`
	DeltaRecovered        string `json:"delta_recovered"`
	Recovered             string `json:"recovered"`
	RecoveredPercent      string `json:"recovered_percent"`
	DeltaRecoveredPercent string `json:"delta_recovered_percent"`
	DeltaActive           string `json:"delta_active"`
	Active                string `json:"active"`
	ActivePercent         string `json:"active_percent"`
	DeltaActivePercent    string `json:"delta_active_percent"`
}
