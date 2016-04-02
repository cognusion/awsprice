package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/spf13/cast"
	"io/ioutil"
	"os"
	"strings"
)

type Json map[string]interface{}

type ReservedPrice struct {
	Available       bool
	Upfront         float64
	Hourly          float64
	AmortizedHourly float64
	years           float64
}

func (r *ReservedPrice) Set(upfront, hourly, years float64) {
	r.Available = true
	r.Upfront = upfront
	r.Hourly = hourly
	r.years = years
	r.AmortizedHourly = hourly + (upfront / (yearly * years))
}

func (r *ReservedPrice) ToString(multiplier float64) string {
	if r.Available {
		return fmt.Sprintf("%f", r.AmortizedHourly*multiplier)
	} else {
		return "N/A"
	}
}

func (r *ReservedPrice) Savings(cost float64) (perc float64) {
	if r.Available {
		rate := cost - r.AmortizedHourly
		perc = (rate / cost) * 100
	}

	return
}

type SkuPrice struct {
	OnDemand     float64
	OnePartial   ReservedPrice
	OneNo        ReservedPrice
	OneAll       ReservedPrice
	ThreePartial ReservedPrice
	ThreeAll     ReservedPrice
}

func (s *SkuPrice) HeaderToString() string {
	return fmt.Sprintf("OnDemand, 1year Partial Upfront, 1year No Upfront, 1year All Upfront, 3year Partial Upfront, 3year All Upfront")
}

func (s *SkuPrice) ToString(multiplier float64) string {
	return fmt.Sprintf("%f,%s,%s,%s,%s,%s", s.OnDemand*multiplier,
		s.OnePartial.ToString(multiplier), s.OneNo.ToString(multiplier),
		s.OneAll.ToString(multiplier), s.ThreePartial.ToString(multiplier),
		s.ThreeAll.ToString(multiplier))
}

type Sku struct {
	Id         string
	Attributes map[string]string
	Pricing    SkuPrice
}

func (s *Sku) PrintHeader() {
	fmt.Printf("Instance Type, %s\n", s.Pricing.HeaderToString())
}

func (s *Sku) Print(multiplier float64) {
	fmt.Printf("%s,%s\n", s.Attributes["instanceType"], s.Pricing.ToString(multiplier))
}

var (
	TermsReserved Json
	TermsDemand   Json

	daily   float64 = 24
	weekly  float64 = daily * 7
	monthly float64 = daily * 30
	yearly  float64 = daily * 365
)

func NewSku(name string, j Json) (s Sku) {

	var p SkuPrice

	// Find the OnDemand price
	p.OnDemand = cast.ToFloat64(cast.ToStringMapString(getFirstElement(cast.ToStringMap(getFirstElement(cast.ToStringMap(TermsDemand[name]))["priceDimensions"]))["pricePerUnit"])["USD"])

	// Grab all the reserved prices
	getReservedPrices(name, &p)

	// Grab the attributes
	attributes := cast.ToStringMapString(j["attributes"])
	attributes["productFamily"] = cast.ToString(j["productFamily"])

	s = Sku{
		Id:         name,
		Attributes: attributes,
		Pricing:    p,
	}

	return
}

func main() {

	var (
		filePath             string
		offerCode            string
		location             string
		locations            []string
		productFamily        string
		instanceFamily       string
		instanceType         string
		instanceTypes        []string
		operatingSystem      string
		operatingSystems     []string
		tenancy              string
		tenancies            []string
		timeunit             string
		timeMultiplier       float64
		ondemand             bool
		onepartial           bool
		onenone              bool
		oneall               bool
		threepartial         bool
		threeall             bool
		csvout               bool
		listlocations        bool
		listinstancefamilies bool
		listproductfamilies  bool
	)

	flag.StringVar(&filePath, "file", "", "The path to the offer file (You should not need to set this)")
	flag.StringVar(&OFFERINDEX, "offerindex", OFFERINDEX, "The URI of the offer index (You should not need to change this)")
	flag.StringVar(&offerCode, "offercode", "AmazonEC2", "If 'file' is empty, this is the offer file that is downloaded")
	flag.StringVar(&location, "location", "US East (N. Virginia)", "The location to filter on")
	flag.StringVar(&productFamily, "family", "Compute Instance", "The product family to filter on")
	flag.StringVar(&instanceFamily, "instancefamily", "", "The instance family to filter on")
	flag.StringVar(&instanceType, "instancetype", "", "The instance type to filter on")
	flag.StringVar(&operatingSystem, "os", "Linux", "The operating system to filter on")
	flag.StringVar(&tenancy, "tenancy", "Shared", "The tenancy type to filter on")
	flag.StringVar(&timeunit, "timeunit", "Hourly", "Hourly, Daily, Weekly, Monthly, Yearly")
	flag.BoolVar(&ondemand, "ondemand", true, "Show OnDemand costs")
	flag.BoolVar(&onepartial, "1partial", false, "Show 1year Partial Upfront costs")
	flag.BoolVar(&onenone, "1none", false, "Show 1year No Upfront costs")
	flag.BoolVar(&oneall, "1all", false, "Show 1year All Upfront costs")
	flag.BoolVar(&threepartial, "3partial", false, "Show 3year Partial Upfront costs")
	flag.BoolVar(&threeall, "3all", false, "Show 3year All Upfront costs")
	flag.BoolVar(&csvout, "csvout", false, "Output in CSV format. All costs will be exported")
	flag.BoolVar(&listlocations, "listlocations", false, "List all of the locations")
	flag.BoolVar(&listinstancefamilies, "listinstancefamilies", false, "List all of the instance families")
	flag.BoolVar(&listproductfamilies, "listproductfamilies", false, "List all of the product families")
	flag.Parse()

	// Splitsies
	locations = strings.Split(location, ",")
	instanceTypes = strings.Split(instanceType, ",")
	operatingSystems = strings.Split(operatingSystem, ",")
	tenancies = strings.Split(tenancy, ",")

	// Set the hour count
	timeMultiplier = 1
	switch timeunit {
	case "Daily":
		timeMultiplier = daily
	case "Weekly":
		timeMultiplier = weekly
	case "Monthly":
		timeMultiplier = monthly
	case "Yearly":
		timeMultiplier = yearly
	}

	var j Json
	if filePath == "" {
		// if exists offerCode.json, just read it
		offerFile := offerCode + ".json"

		if _, err := os.Stat(offerFile); os.IsNotExist(err) {
			// Doesn't exist...

			// Grab the index
			fmt.Printf("No cached file for %s\nDownloading offer index...\n", offerCode)
			offerIndex, err := http2json(OFFERHOST + OFFERINDEX)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Download the offer file. This will take a while, so we'll be using a
			// a progress bar
			fmt.Printf("Offer index: %s\nDownloading offer file for %s\n", cast.ToString(offerIndex["publicationDate"]), offerCode)

			offer := cast.ToStringMapString(cast.ToStringMap(offerIndex["offers"])[offerCode])["currentVersionUrl"]
			err = http2file(OFFERHOST+offer, offerFile, true)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		// Load the cached file
		j = loadJsonFile(offerFile)
		fmt.Printf("Offer file: %s\n", cast.ToString(j["publicationDate"]))

	} else {
		// Load the specified file
		j = loadJsonFile(filePath)
	}

	products := cast.ToStringMap(j["products"])
	TermsReserved = cast.ToStringMap(cast.ToStringMap(j["terms"])["Reserved"])
	TermsDemand = cast.ToStringMap(cast.ToStringMap(j["terms"])["OnDemand"])

	if listlocations || listinstancefamilies || listproductfamilies {

		stuff := make(map[string]bool)

		for _, v := range products {
			p := cast.ToStringMap(v)
			a := cast.ToStringMapString(p["attributes"])

			if listproductfamilies {
				pfam := cast.ToString(p["productFamily"])
				if pfam != "" {
					stuff[pfam] = true
				}
			}

			if listinstancefamilies {
				if a["instanceFamily"] != "" {
					stuff[a["instanceFamily"]] = true
				}
			}

			if listlocations {
				if a["location"] != "" {
					stuff[a["location"]] = true
				}
			}
		}

		for k, _ := range stuff {
			fmt.Println(k)
		}
		
		os.Exit(0)

	}

	// No listings, run the op
	firstPrint := true
	for k, v := range products {
		p := cast.ToStringMap(v)
		if productFamily != "" && productFamily != cast.ToString(p["productFamily"]) {
			// Not it
			continue
		}

		// Filter on attributes
		a := cast.ToStringMapString(p["attributes"])

		if instanceFamily != "" && instanceFamily != a["instanceFamily"] {
			// Not these
			continue
		}

		if location != "" {
			found := false
			for _, l := range locations {
				if l == a["location"] {
					found = true
					break
				}
			}
			if found == false {
				// Not here
				continue
			}
		}

		if instanceType != "" {
			found := false
			for _, l := range instanceTypes {
				if l == a["instanceType"] {
					found = true
					break
				}
			}
			if found == false {
				// Wrong porridge
				continue
			}
		}

		if operatingSystem != "" {
			found := false
			for _, l := range operatingSystems {
				if l == a["operatingSystem"] {
					found = true
					break
				}
			}
			if found == false {
				// Wrong droid
				continue
			}
		}

		if tenancy != "" {
			found := false
			for _, l := range tenancies {
				if l == a["tenancy"] {
					found = true
					break
				}
			}
			if found == false {
				// Wrong vert
				continue
			}
		}

		// POST: a valid sku
		s := NewSku(k, p)

		// Print all the things!
		if csvout {
			if firstPrint {
				s.PrintHeader()
				firstPrint = false
			}

			s.Print(timeMultiplier)
		} else {

			fmt.Printf("%s (%s) %s RAM: %s VCPU: %s\n", s.Attributes["instanceType"], s.Attributes["operatingSystem"], s.Attributes["location"], s.Attributes["memory"], s.Attributes["vcpu"])

			if ondemand {
				fmt.Printf("\tOnDemand $%f\n", s.Pricing.OnDemand*timeMultiplier)
			}
			if onepartial {
				fmt.Printf("\t1year Partial Upfront: (%f%%)\n\t\tUpfront $%f\n\t\t%s $%f\n\t\t%s (amortized) $%f\n",
					s.Pricing.OnePartial.Savings(s.Pricing.OnDemand), s.Pricing.OnePartial.Upfront,
					timeunit, s.Pricing.OnePartial.Hourly*timeMultiplier, timeunit, s.Pricing.OnePartial.AmortizedHourly*timeMultiplier)
			}
			if onenone {
				fmt.Printf("\t1year No Upfront: (%f%%)\n\t\tUpfront $%f\n\t\t%s $%f\n\t\t%s (amortized) $%f\n",
					s.Pricing.OneNo.Savings(s.Pricing.OnDemand), s.Pricing.OneNo.Upfront, timeunit, s.Pricing.OneNo.Hourly*timeMultiplier,
					timeunit, s.Pricing.OneNo.AmortizedHourly*timeMultiplier)
			}
			if oneall {
				fmt.Printf("\t1year All Upfront: (%f%%)\n\t\tUpfront $%f\n\t\t%s $%f\n\t\t%s (amortized) $%f\n",
					s.Pricing.OneAll.Savings(s.Pricing.OnDemand), s.Pricing.OneAll.Upfront,
					timeunit, s.Pricing.OneAll.Hourly*timeMultiplier, timeunit, s.Pricing.OneAll.AmortizedHourly*timeMultiplier)
			}
			if threepartial {
				fmt.Printf("\t3year Partial Upfront: (%f%%)\n\t\tUpfront $%f\n\t\t%s $%f\n\t\t%s (amortized) $%f\n",
					s.Pricing.ThreePartial.Savings(s.Pricing.OnDemand), s.Pricing.ThreePartial.Upfront, timeunit, s.Pricing.ThreePartial.Hourly*timeMultiplier,
					timeunit, s.Pricing.ThreePartial.AmortizedHourly*timeMultiplier)
			}
			if threeall {
				fmt.Printf("\t3year All Upfront: (%f%%)\n\t\tUpfront $%f\n\t\t%s $%f\n\t\t%s (amortized) $%f\n",
					s.Pricing.ThreeAll.Savings(s.Pricing.OnDemand), s.Pricing.ThreeAll.Upfront, timeunit, s.Pricing.ThreeAll.Hourly*timeMultiplier,
					timeunit, s.Pricing.ThreeAll.AmortizedHourly*timeMultiplier)
			}
		}

	}
}

// A helper function to simply grab the "first" element of a Json doc
func getFirstElement(a Json) (j Json) {
	for _, v := range a {
		j = cast.ToStringMap(v)
		break
	}
	return
}

// A function to mask the ugliness of diving a couple levels in to the
// reserved offer file, which has a horrific schema. This could certainly
// be done better, but is sufficient for a short-running program
func getReservedPrices(sku string, skuPrice *SkuPrice) {

	for _, v := range cast.ToStringMap(TermsReserved[sku]) {
		vm := cast.ToStringMap(v)
		terms := cast.ToStringMapString(vm["termAttributes"])
		length := terms["LeaseContractLength"]
		option := terms["PurchaseOption"]

		var (
			upfront float64
			hourly  float64
		)

		// Diving into the priceDimensions to grab the prices
		for _, v2 := range cast.ToStringMap(vm["priceDimensions"]) {
			v2m := cast.ToStringMap(v2)
			if cast.ToString(v2m["unit"]) == "Quantity" {
				upfront = cast.ToFloat64(cast.ToStringMap(v2m["pricePerUnit"])["USD"])
			} else if cast.ToString(v2m["unit"]) == "Hrs" {
				hourly = cast.ToFloat64(cast.ToStringMap(v2m["pricePerUnit"])["USD"])
			}
		}

		// Cascading switch on the term length and payment option
		switch length {
		case "1yr":
			switch option {
			case "Partial Upfront":
				skuPrice.OnePartial.Set(upfront, hourly, 1)
			case "No Upfront":
				skuPrice.OneNo.Set(upfront, hourly, 1)
			case "All Upfront":
				skuPrice.OneAll.Set(upfront, hourly, 1)
			}
		case "3yr":
			switch option {
			case "Partial Upfront":
				skuPrice.ThreePartial.Set(upfront, hourly, 3)
			case "All Upfront":
				skuPrice.ThreeAll.Set(upfront, hourly, 3)
			}
		}
	}
}

// Loads a JSON file and returns a Json type
func loadJsonFile(filePath string) (j Json) {

	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading JSON file '%s': %s\n", filePath, err)
		os.Exit(1)
	}

	err = json.Unmarshal(buf, &j)
	if err != nil {
		fmt.Printf("Error parsing JSON file '%s': %s\n", filePath, err)
		os.Exit(1)

	}
	return
}
