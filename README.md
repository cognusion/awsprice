# awsprice
Grabs and caches the AWS pricelist and gives you the prices you request

Gotchas
-------

Only EC2 "Compute Instance" types are supported at this time.

The "location" is not formatted like other AWS services (thanks, Amazon). Use --listlocations
to see all of the locations in the downloaded/cached offer file

Results are not sorted, and their order means nothing.

There's very little sanity checking of your commands. You can specify outrageous things,
and get back results that aren't actually what you asked for.

Basics
------

```bash
go get -u github.com/cognusion/awsprice

$ awsprice --help
Usage of awsprice:
  -1all
    	Show 1year All Upfront costs
  -1none
    	Show 1year No Upfront costs
  -1partial
    	Show 1year Partial Upfront costs
  -3all
    	Show 3year All Upfront costs
  -3partial
    	Show 3year Partial Upfront costs
  -csvout
    	Output in CSV format. All costs will be exported
  -family string
    	The product family to filter on (default "Compute Instance")
  -file string
    	The path to the offer file (You should not need to set this)
  -instancefamily string
    	The instance family to filter on
  -instancetype string
    	The instance type to filter on
  -listinstancefamilies
    	List all of the instance families
  -listlocations
    	List all of the locations
  -listproductfamilies
    	List all of the product families
  -location string
    	The location to filter on (default "US East (N. Virginia)")
  -offercode string
    	If 'file' is empty, this is the offer file that is downloaded (default "AmazonEC2")
  -offerindex string
    	The URI of the offer index (You should not need to change this) (default "/offers/v1.0/aws/index.json")
  -ondemand
    	Show OnDemand costs (default true)
  -os string
    	The operating system to filter on (default "Linux")
  -tenancy string
    	The tenancy type to filter on (default "Shared")
  -timeunit string
    	Hourly, Daily, Weekly, Monthly, Yearly (default "Hourly")
 
 ```
 
Examples
--------
 
List all the locations
```bash
$ awsprice --listlocations
Offer file: 2016-01-26T00:17:08Z
Asia Pacific (Tokyo)
AWS GovCloud (US)
EU (Ireland)
South America (Sao Paulo)
US East (N. Virginia)
Asia Pacific (Seoul)
Asia Pacific (Singapore)
US West (Oregon)
Asia Pacific (Sydney)
EU (Frankfurt)
US West (N. California)
```

Compare the OnDemand costs only, Monthly cost calculated, for r3.large and m4.large instances
 ```bash
$ awsprice --instancetype r3.large,m4.large --timeunit Monthly
Offer file: 2016-01-26T00:17:08Z
r3.large (Linux) US East (N. Virginia) RAM: 15.25 GiB VCPU: 2
	OnDemand $119.520000
m4.large (Linux) US East (N. Virginia) RAM: 8 GiB VCPU: 2
	OnDemand $86.400000
```

Compare the OnDemand costs only, Monthly cost calculated, for r3.large and m4.large instances,
between free Linux, RHEL-license, and SUSE-license
```bash
$ awsprice --instancetype r3.large,m4.large --timeunit Monthly --os Linux,RHEL,SUSE
Offer file: 2016-01-26T00:17:08Z
r3.large (RHEL) US East (N. Virginia) RAM: 15.25 GiB VCPU: 2
	OnDemand $162.720000
m4.large (RHEL) US East (N. Virginia) RAM: 8 GiB VCPU: 2
	OnDemand $129.600000
r3.large (Linux) US East (N. Virginia) RAM: 15.25 GiB VCPU: 2
	OnDemand $119.520000
m4.large (Linux) US East (N. Virginia) RAM: 8 GiB VCPU: 2
	OnDemand $86.400000
r3.large (SUSE) US East (N. Virginia) RAM: 15.25 GiB VCPU: 2
	OnDemand $191.520000
m4.large (SUSE) US East (N. Virginia) RAM: 8 GiB VCPU: 2
	OnDemand $158.400000
```

All of the "Memory optimized" instances, OnDemand costs only, Monthly cost calculated
```bash
 $ awsprice --instancefamily "Memory optimized" --timeunit Monthly
Offer file: 2016-01-26T00:17:08Z
r3.large (Linux) US East (N. Virginia) RAM: 15.25 GiB VCPU: 2
	OnDemand $119.520000
r3.2xlarge (Linux) US East (N. Virginia) RAM: 61 GiB VCPU: 8
	OnDemand $478.800000
cr1.8xlarge (Linux) US East (N. Virginia) RAM: 244 GiB VCPU: 32
	OnDemand $2520.000000
r3.xlarge (Linux) US East (N. Virginia) RAM: 30.5 GiB VCPU: 4
	OnDemand $239.760000
m2.2xlarge (Linux) US East (N. Virginia) RAM: 34.2 GiB VCPU: 4
	OnDemand $352.800000
r3.4xlarge (Linux) US East (N. Virginia) RAM: 122 GiB VCPU: 16
	OnDemand $957.600000
r3.8xlarge (Linux) US East (N. Virginia) RAM: 244 GiB VCPU: 32
	OnDemand $1915.200000
m2.4xlarge (Linux) US East (N. Virginia) RAM: 68.4 GiB VCPU: 8
	OnDemand $705.600000
m2.xlarge (Linux) US East (N. Virginia) RAM: 17.1 GiB
```

All of the "Memory optimized" instances, OnDemand vs 1year Partial vs 1 year Full 
reserved costs, Monthly cost calculated

```bash
$ awsprice --instancefamily "Memory optimized" --timeunit Monthly --1partial --1all
Offer file: 2016-01-26T00:17:08Z
m2.4xlarge (Linux) US East (N. Virginia) RAM: 68.4 GiB VCPU: 8
	OnDemand $705.600000
	1year Partial Upfront: (61.305097%)
		Upfront $1894.000000
		Monthly $117.360000
		Monthly (amortized) $273.031233
	1year All Upfront: (62.084149%)
		Upfront $3255.000000
		Monthly $0.000000
		Monthly (amortized) $267.534247
m2.xlarge (Linux) US East (N. Virginia) RAM: 17.1 GiB VCPU: 2
	OnDemand $176.400000
	1year Partial Upfront: (61.226354%)
		Upfront $473.000000
		Monthly $29.520000
		Monthly (amortized) $68.396712
	1year All Upfront: (62.025906%)
		Upfront $815.000000
		Monthly $0.000000
		Monthly (amortized) $66.986301
r3.8xlarge (Linux) US East (N. Virginia) RAM: 244 GiB VCPU: 32
	OnDemand $1915.200000
	1year Partial Upfront: (45.838397%)
		Upfront $8223.000000
		Monthly $361.440000
		Monthly (amortized) $1037.303014
	1year All Upfront: (46.939266%)
		Upfront $12364.000000
		Monthly $0.000000
		Monthly (amortized) $1016.219178
cr1.8xlarge (Linux) US East (N. Virginia) RAM: 244 GiB VCPU: 32
	OnDemand $2520.000000
	1year Partial Upfront: (76.451402%)
		Upfront $7220.000000
		Monthly $0.000000
		Monthly (amortized) $593.424658
	1year All Upfront: (59.562948%)
		Upfront $12398.000000
		Monthly $0.000000
		Monthly (amortized) $1019.013699
r3.2xlarge (Linux) US East (N. Virginia) RAM: 61 GiB VCPU: 8
	OnDemand $478.800000
	1year Partial Upfront: (45.909294%)
		Upfront $2056.000000
		Monthly $90.000000
		Monthly (amortized) $258.986301
	1year All Upfront: (46.956432%)
		Upfront $3090.000000
		Monthly $0.000000
		Monthly (amortized) $253.972603
r3.xlarge (Linux) US East (N. Virginia) RAM: 30.5 GiB VCPU: 4
	OnDemand $239.760000
	1year Partial Upfront: (45.840361%)
		Upfront $1028.000000
		Monthly $45.360000
		Monthly (amortized) $129.853151
	1year All Upfront: (47.036077%)
		Upfront $1545.000000
		Monthly $0.000000
		Monthly (amortized) $126.986301
r3.4xlarge (Linux) US East (N. Virginia) RAM: 122 GiB VCPU: 16
	OnDemand $957.600000
	1year Partial Upfront: (45.834106%)
		Upfront $4112.000000
		Monthly $180.720000
		Monthly (amortized) $518.692603
	1year All Upfront: (46.939266%)
		Upfront $6182.000000
		Monthly $0.000000
		Monthly (amortized) $508.109589
m2.2xlarge (Linux) US East (N. Virginia) RAM: 34.2 GiB VCPU: 4
	OnDemand $352.800000
	1year Partial Upfront: (61.203057%)
		Upfront $947.000000
		Monthly $59.040000
		Monthly (amortized) $136.875616
	1year All Upfront: (61.979312%)
		Upfront $1632.000000
		Monthly $0.000000
		Monthly (amortized) $134.136986
r3.large (Linux) US East (N. Virginia) RAM: 15.25 GiB VCPU: 2
	OnDemand $119.520000
	1year Partial Upfront: (45.978434%)
		Upfront $514.000000
		Monthly $22.320000
		Monthly (amortized) $64.566575
	1year All Upfront: (46.910931%)
		Upfront $772.000000
		Monthly $0.000000
		Monthly (amortized) $63.452055
		
```

Compare the price differences, between US East (N. Virginia) and EU (Frankfurt), for an r3.xlarge, with a couple reservation types
```bash
$ awsprice --timeunit Monthly --instancetype r3.xlarge --1partial --1all --location "US East (N. Virginia),EU (Frankfurt)"
Offer file: 2016-01-26T00:17:08Z
r3.xlarge (Linux) US East (N. Virginia) RAM: 30.5 GiB VCPU: 4
	OnDemand $239.760000
	1year Partial Upfront: (45.840361%)
		Upfront $1028.000000
		Monthly $45.360000
		Monthly (amortized) $129.853151
	1year All Upfront: (47.036077%)
		Upfront $1545.000000
		Monthly $0.000000
		Monthly (amortized) $126.986301
r3.xlarge (Linux) EU (Frankfurt) RAM: 30.5 GiB VCPU: 4
	OnDemand $288.000000
	1year Partial Upfront: (37.034247%)
		Upfront $1488.000000
		Monthly $59.040000
		Monthly (amortized) $181.341370
	1year All Upfront: (38.413242%)
		Upfront $2158.000000
		Monthly $0.000000
		Monthly (amortized) $177.369863
		
```