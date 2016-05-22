package jarvisbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/tucnak/telebot"
)

const openExchangeRateURL = "https://openexchangerates.org/api/latest.json?app_id="

// Exchange is used to perform an exchange rate conversion.
func (j *JarvisBot) Exchange(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.SendMessage(msg.Chat, "/xchg: Do an exchange rate conversion\nHere are some commands to try: \n* 10 sgd in usd\n* 100 vnd to sgd\n* 21 usd how much arr?\n\n\U0001F4A1 You could also use this format for faster results:\n/x 10 sgd in usd", so)
		return
	}

	amount, fromCurr, toCurr := parseArgs(msg.Args)
	if amount == 0.0 || fromCurr == "" || toCurr == "" {
		j.SendMessage(msg.Chat, "I didn't understand that. Here are some commands to try: \n/xchg 10 sgd in usd\n/xchg 100 vnd to sgd\n/xchg 21 usd how much arr?", nil)
		return
	}

	fromCurrRate, toCurrRate, err := j.getRatesFromDB(fromCurr, toCurr)
	if err != nil {
		j.log.Printf("[%s] problem with retrieving rates: %s", time.Now().Format(time.RFC3339), err)
		return
	}

	res := amount * 1 / fromCurrRate * toCurrRate
	displayRate := 1 / fromCurrRate * toCurrRate

	strDisplayRate := strconv.FormatFloat(displayRate, 'f', 5, 64)
	fmtAmount := strconv.FormatFloat(res, 'f', 2, 64)

	j.SendMessage(msg.Chat, "\U0001F4B8 "+fromCurr+" to "+toCurr+"\nRate: 1.00 : "+strDisplayRate+"\n"+strconv.FormatFloat(amount, 'f', 2, 64)+" "+fromCurr+" = "+fmtAmount+" "+toCurr, nil)
}

// Retrieve rates from DB.
func (j *JarvisBot) getRatesFromDB(fromCurr, toCurr string) (float64, float64, error) {
	if j.ratesAreEmpty() {
		j.log.Println("retrieving rates due to an empty database")
		err := j.RetrieveAndSaveExchangeRates()
		if err != nil {
			return 0.0, 0.0, err
		}
	}

	var fromCurrRate, toCurrRate float64
	err := j.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket(exchange_rate_bucket_name)
		v := b.Get([]byte(fromCurr))
		fromCurrRate, err = strconv.ParseFloat(string(v), 64)
		if err != nil {
			return err
		}

		v = b.Get([]byte(toCurr))
		toCurrRate, err = strconv.ParseFloat(string(v), 64)
		if err != nil {
			return err
		}
		return nil
	})

	return fromCurrRate, toCurrRate, err
}

// RetrieveAndSaveExchangeRates retrieves exchange rates and saves it to DB.
func (j *JarvisBot) RetrieveAndSaveExchangeRates() error {
	rates, err := j.RetrieveExchangeRates()
	if err != nil {
		return err
	}

	err = j.db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket(exchange_rate_bucket_name)
		err := b.Put([]byte("timestamp"), []byte(strconv.Itoa(rates.UnixTimestamp)))
		if err != nil {
			return fmt.Errorf("error saving timestamp %s to db: %s", strconv.Itoa(rates.UnixTimestamp), err)
		}

		for k, v := range rates.Rates {
			err := b.Put([]byte(k), []byte(strconv.FormatFloat(v, 'f', -1, 64)))
			if err != nil {
				return fmt.Errorf("error saving value %s:%f to db: %s", k, v, err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

// Checks to see if the rates database is empty.
func (j *JarvisBot) ratesAreEmpty() bool {
	res := true
	j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(exchange_rate_bucket_name)
		stats := b.Stats()
		if stats.KeyN > 0 {
			res = false
		}
		return nil
	})
	return res
}

// Rates stores currency exchange rates
type Rates struct {
	UnixTimestamp int                `json:"timestamp"`
	Base          string             `json:"base"`
	Rates         map[string]float64 `json:"rates"`
}

// Retrieves exchange rates from the OpenExchangeAPI.
func (j *JarvisBot) RetrieveExchangeRates() (*Rates, error) {
	if j.keys.OpenExchangeAPIKey == "" {
		err := fmt.Errorf("no open exchange api key!")
		return nil, err
	}
	resp, err := http.Get(openExchangeRateURL + j.keys.OpenExchangeAPIKey)
	if err != nil {
		return nil, err
	}

	jsonBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var rates Rates
	err = json.Unmarshal(jsonBody, &rates)
	if err != nil {
		return nil, err
	}

	return &rates, nil
}

// Helper functions
func parseArgs(args []string) (amount float64, fromCurr, toCurr string) {
	amount = 0.0
	fromCurr, toCurr = "", ""

	for _, a := range args {
		s := strings.ToUpper(a)

		if currencyCode[s] != "" {
			if fromCurr == "" {
				fromCurr = currencyCode[s]
			} else if toCurr == "" {
				toCurr = currencyCode[s]
			}
		} else {
			f, err := strconv.ParseFloat(a, 64)
			// We take the first number in the string only.
			if err == nil && amount == 0.0 && f > 0 {
				amount = f
			}
		}
	}
	if toCurr == "" && fromCurr != "" {
		toCurr = "SGD"
	}
	if amount == 0.0 {
		amount = 1.0
	}
	return amount, fromCurr, toCurr
}

var currencyCode = map[string]string{
	"RINGGIT":  "MYR",
	"BAHT":     "THB",
	"POUNDS":   "GBP",
	"POUND":    "GBP",
	"RUPIAH":   "IDR",
	"SING":     "SGD",
	"SG":       "SGD",
	"RMB":      "CNY",
	"RENMINBI": "CNY",
	"RM":       "MYR",
	"YEN":      "JPY",
	"YUAN":     "CNY",
	"EURO":     "EUR",
	"EUROS":    "EUR",
	"DONG":     "VND",
	"AED":      "AED",
	"AFN":      "AFN",
	"ALL":      "ALL",
	"AMD":      "AMD",
	"ANG":      "ANG",
	"AOA":      "AOA",
	"ARS":      "ARS",
	"AUD":      "AUD",
	"AWG":      "AWG",
	"AZN":      "AZN",
	"BAM":      "BAM",
	"BBD":      "BBD",
	"BDT":      "BDT",
	"BGN":      "BGN",
	"BHD":      "BHD",
	"BIF":      "BIF",
	"BMD":      "BMD",
	"BND":      "BND",
	"BOB":      "BOB",
	"BRL":      "BRL",
	"BSD":      "BSD",
	"BTC":      "BTC",
	"BTN":      "BTN",
	"BWP":      "BWP",
	"BYR":      "BYR",
	"BZD":      "BZD",
	"CAD":      "CAD",
	"CDF":      "CDF",
	"CHF":      "CHF",
	"CLF":      "CLF",
	"CLP":      "CLP",
	"CNY":      "CNY",
	"COP":      "COP",
	"CRC":      "CRC",
	"CUC":      "CUC",
	"CUP":      "CUP",
	"CVE":      "CVE",
	"CZK":      "CZK",
	"DJF":      "DJF",
	"DKK":      "DKK",
	"DOP":      "DOP",
	"DZD":      "DZD",
	"EEK":      "EEK",
	"EGP":      "EGP",
	"ERN":      "ERN",
	"ETB":      "ETB",
	"EUR":      "EUR",
	"FJD":      "FJD",
	"FKP":      "FKP",
	"GBP":      "GBP",
	"GEL":      "GEL",
	"GGP":      "GGP",
	"GHS":      "GHS",
	"GIP":      "GIP",
	"GMD":      "GMD",
	"GNF":      "GNF",
	"GTQ":      "GTQ",
	"GYD":      "GYD",
	"HKD":      "HKD",
	"HNL":      "HNL",
	"HRK":      "HRK",
	"HTG":      "HTG",
	"HUF":      "HUF",
	"IDR":      "IDR",
	"ILS":      "ILS",
	"IMP":      "IMP",
	"INR":      "INR",
	"IQD":      "IQD",
	"IRR":      "IRR",
	"ISK":      "ISK",
	"JEP":      "JEP",
	"JMD":      "JMD",
	"JOD":      "JOD",
	"JPY":      "JPY",
	"KES":      "KES",
	"KGS":      "KGS",
	"KHR":      "KHR",
	"KMF":      "KMF",
	"KPW":      "KPW",
	"KRW":      "KRW",
	"KWD":      "KWD",
	"KYD":      "KYD",
	"KZT":      "KZT",
	"LAK":      "LAK",
	"LBP":      "LBP",
	"LKR":      "LKR",
	"LRD":      "LRD",
	"LSL":      "LSL",
	"LTL":      "LTL",
	"LVL":      "LVL",
	"LYD":      "LYD",
	"MAD":      "MAD",
	"MDL":      "MDL",
	"MGA":      "MGA",
	"MKD":      "MKD",
	"MMK":      "MMK",
	"MNT":      "MNT",
	"MOP":      "MOP",
	"MRO":      "MRO",
	"MTL":      "MTL",
	"MUR":      "MUR",
	"MVR":      "MVR",
	"MWK":      "MWK",
	"MXN":      "MXN",
	"MYR":      "MYR",
	"MZN":      "MZN",
	"NAD":      "NAD",
	"NGN":      "NGN",
	"NIO":      "NIO",
	"NOK":      "NOK",
	"NPR":      "NPR",
	"NZD":      "NZD",
	"OMR":      "OMR",
	"PAB":      "PAB",
	"PEN":      "PEN",
	"PGK":      "PGK",
	"PHP":      "PHP",
	"PKR":      "PKR",
	"PLN":      "PLN",
	"PYG":      "PYG",
	"QAR":      "QAR",
	"RON":      "RON",
	"RSD":      "RSD",
	"RUB":      "RUB",
	"RWF":      "RWF",
	"SAR":      "SAR",
	"SBD":      "SBD",
	"SCR":      "SCR",
	"SDG":      "SDG",
	"SEK":      "SEK",
	"SGD":      "SGD",
	"SHP":      "SHP",
	"SLL":      "SLL",
	"SOS":      "SOS",
	"SRD":      "SRD",
	"STD":      "STD",
	"SVC":      "SVC",
	"SYP":      "SYP",
	"SZL":      "SZL",
	"THB":      "THB",
	"TJS":      "TJS",
	"TMT":      "TMT",
	"TND":      "TND",
	"TOP":      "TOP",
	"TRY":      "TRY",
	"TTD":      "TTD",
	"TWD":      "TWD",
	"TZS":      "TZS",
	"UAH":      "UAH",
	"UGX":      "UGX",
	"USD":      "USD",
	"UYU":      "UYU",
	"UZS":      "UZS",
	"VEF":      "VEF",
	"VND":      "VND",
	"VUV":      "VUV",
	"WST":      "WST",
	"XAF":      "XAF",
	"XAG":      "XAG",
	"XAU":      "XAU",
	"XCD":      "XCD",
	"XDR":      "XDR",
	"XOF":      "XOF",
	"XPD":      "XPD",
	"XPF":      "XPF",
	"XPT":      "XPT",
	"YER":      "YER",
	"ZAR":      "ZAR",
	"ZMK":      "ZMK",
	"ZMW":      "ZMW",
	"ZWL":      "ZWL",
}

var currencyName = map[string]string{
	"AED": "United Arab Emirates Dirham",
	"AFN": "Afghan Afghani",
	"ALL": "Albanian Lek",
	"AMD": "Armenian Dram",
	"ANG": "Netherlands Antillean Guilder",
	"AOA": "Angolan Kwanza",
	"ARS": "Argentine Peso",
	"AUD": "Australian Dollar",
	"AWG": "Aruban Florin",
	"AZN": "Azerbaijani Manat",
	"BAM": "Bosnia-Herzegovina Convertible Mark",
	"BBD": "Barbadian Dollar",
	"BDT": "Bangladeshi Taka",
	"BGN": "Bulgarian Lev",
	"BHD": "Bahraini Dinar",
	"BIF": "Burundian Franc",
	"BMD": "Bermudan Dollar",
	"BND": "Brunei Dollar",
	"BOB": "Bolivian Boliviano",
	"BRL": "Brazilian Real",
	"BSD": "Bahamian Dollar",
	"BTC": "Bitcoin",
	"BTN": "Bhutanese Ngultrum",
	"BWP": "Botswanan Pula",
	"BYR": "Belarusian Ruble",
	"BZD": "Belize Dollar",
	"CAD": "Canadian Dollar",
	"CDF": "Congolese Franc",
	"CHF": "Swiss Franc",
	"CLF": "Chilean Unit of Account (UF)",
	"CLP": "Chilean Peso",
	"CNY": "Chinese Yuan",
	"COP": "Colombian Peso",
	"CRC": "Costa Rican Colón",
	"CUC": "Cuban Convertible Peso",
	"CUP": "Cuban Peso",
	"CVE": "Cape Verdean Escudo",
	"CZK": "Czech Republic Koruna",
	"DJF": "Djiboutian Franc",
	"DKK": "Danish Krone",
	"DOP": "Dominican Peso",
	"DZD": "Algerian Dinar",
	"EEK": "Estonian Kroon",
	"EGP": "Egyptian Pound",
	"ERN": "Eritrean Nakfa",
	"ETB": "Ethiopian Birr",
	"EUR": "Euro",
	"FJD": "Fijian Dollar",
	"FKP": "Falkland Islands Pound",
	"GBP": "British Pound Sterling",
	"GEL": "Georgian Lari",
	"GGP": "Guernsey Pound",
	"GHS": "Ghanaian Cedi",
	"GIP": "Gibraltar Pound",
	"GMD": "Gambian Dalasi",
	"GNF": "Guinean Franc",
	"GTQ": "Guatemalan Quetzal",
	"GYD": "Guyanaese Dollar",
	"HKD": "Hong Kong Dollar",
	"HNL": "Honduran Lempira",
	"HRK": "Croatian Kuna",
	"HTG": "Haitian Gourde",
	"HUF": "Hungarian Forint",
	"IDR": "Indonesian Rupiah",
	"ILS": "Israeli New Sheqel",
	"IMP": "Manx pound",
	"INR": "Indian Rupee",
	"IQD": "Iraqi Dinar",
	"IRR": "Iranian Rial",
	"ISK": "Icelandic Króna",
	"JEP": "Jersey Pound",
	"JMD": "Jamaican Dollar",
	"JOD": "Jordanian Dinar",
	"JPY": "Japanese Yen",
	"KES": "Kenyan Shilling",
	"KGS": "Kyrgystani Som",
	"KHR": "Cambodian Riel",
	"KMF": "Comorian Franc",
	"KPW": "North Korean Won",
	"KRW": "South Korean Won",
	"KWD": "Kuwaiti Dinar",
	"KYD": "Cayman Islands Dollar",
	"KZT": "Kazakhstani Tenge",
	"LAK": "Laotian Kip",
	"LBP": "Lebanese Pound",
	"LKR": "Sri Lankan Rupee",
	"LRD": "Liberian Dollar",
	"LSL": "Lesotho Loti",
	"LTL": "Lithuanian Litas",
	"LVL": "Latvian Lats",
	"LYD": "Libyan Dinar",
	"MAD": "Moroccan Dirham",
	"MDL": "Moldovan Leu",
	"MGA": "Malagasy Ariary",
	"MKD": "Macedonian Denar",
	"MMK": "Myanma Kyat",
	"MNT": "Mongolian Tugrik",
	"MOP": "Macanese Pataca",
	"MRO": "Mauritanian Ouguiya",
	"MTL": "Maltese Lira",
	"MUR": "Mauritian Rupee",
	"MVR": "Maldivian Rufiyaa",
	"MWK": "Malawian Kwacha",
	"MXN": "Mexican Peso",
	"MYR": "Malaysian Ringgit",
	"MZN": "Mozambican Metical",
	"NAD": "Namibian Dollar",
	"NGN": "Nigerian Naira",
	"NIO": "Nicaraguan Córdoba",
	"NOK": "Norwegian Krone",
	"NPR": "Nepalese Rupee",
	"NZD": "New Zealand Dollar",
	"OMR": "Omani Rial",
	"PAB": "Panamanian Balboa",
	"PEN": "Peruvian Nuevo Sol",
	"PGK": "Papua New Guinean Kina",
	"PHP": "Philippine Peso",
	"PKR": "Pakistani Rupee",
	"PLN": "Polish Zloty",
	"PYG": "Paraguayan Guarani",
	"QAR": "Qatari Rial",
	"RON": "Romanian Leu",
	"RSD": "Serbian Dinar",
	"RUB": "Russian Ruble",
	"RWF": "Rwandan Franc",
	"SAR": "Saudi Riyal",
	"SBD": "Solomon Islands Dollar",
	"SCR": "Seychellois Rupee",
	"SDG": "Sudanese Pound",
	"SEK": "Swedish Krona",
	"SGD": "Singapore Dollar",
	"SHP": "Saint Helena Pound",
	"SLL": "Sierra Leonean Leone",
	"SOS": "Somali Shilling",
	"SRD": "Surinamese Dollar",
	"STD": "São Tomé and Príncipe Dobra",
	"SVC": "Salvadoran Colón",
	"SYP": "Syrian Pound",
	"SZL": "Swazi Lilangeni",
	"THB": "Thai Baht",
	"TJS": "Tajikistani Somoni",
	"TMT": "Turkmenistani Manat",
	"TND": "Tunisian Dinar",
	"TOP": "Tongan Paʻanga",
	"TRY": "Turkish Lira",
	"TTD": "Trinidad and Tobago Dollar",
	"TWD": "New Taiwan Dollar",
	"TZS": "Tanzanian Shilling",
	"UAH": "Ukrainian Hryvnia",
	"UGX": "Ugandan Shilling",
	"USD": "United States Dollar",
	"UYU": "Uruguayan Peso",
	"UZS": "Uzbekistan Som",
	"VEF": "Venezuelan Bolívar Fuerte",
	"VND": "Vietnamese Dong",
	"VUV": "Vanuatu Vatu",
	"WST": "Samoan Tala",
	"XAF": "CFA Franc BEAC",
	"XAG": "Silver (troy ounce)",
	"XAU": "Gold (troy ounce)",
	"XCD": "East Caribbean Dollar",
	"XDR": "Special Drawing Rights",
	"XOF": "CFA Franc BCEAO",
	"XPD": "Palladium Ounce",
	"XPF": "CFP Franc",
	"XPT": "Platinum Ounce",
	"YER": "Yemeni Rial",
	"ZAR": "South African Rand",
	"ZMK": "Zambian Kwacha (pre-2013)",
	"ZMW": "Zambian Kwacha",
	"ZWL": "Zimbabwean Dollar",
}
