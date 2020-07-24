# ruller

[![Build Status](https://travis-ci.org/flaviostutz/ruller.svg?branch=master)](https://travis-ci.org/flaviostutz/ruller)

A simple REST based rule engine in which rules are written in Go

1. You create and register some rules as Go functions in a "main" program and run it
2. You POST to `/rules/[group-name]` along with some JSON body.
3. All rules for that group are processed using the request body.
4. Depending on your implementation, some rules returns data and some rules not.
4. Finally, all rule's results are merged and returned to the REST caller as a JSON.

Ruller works by invoking a bunch of rules with the same input from a single REST call, merging rules outputs and returning the result to the caller. 

You can use this for feature enablement, dynamic systems configuration and other applications where a static key value store wouldn't help you because you need some logic on an input value in order to determine the output.

Checkout some [benchmarks](BENCHMARK.md) we made this far too.

## Example

Checkout the [ruller-sample project](sample).

## Special parameters on POST body

* "_flatten" - true|false. If true, a flat map with all keys returned by all rules, with results merged, will be returned. If false, will return the results with the same tree shape as the rules itself. Defaults to true

* "_keepFirst" - true|false. When using flat map as result, this determines whetever to keep the value from the first or the last rule processed during merge. Default is true

* "_info" - true|false. If true, will add the attribute "_rule" with the name of the rule that generated the node on the result tree (if not using flat map as result). Default to true

## Input parameters used as rules input

* The POST body JSON elements will be converted to a map and used as input parameters
* Additionally, "\_remote\_ip" is set with client remote address

* If you define a geolite2 database using "--geolite2-db", Ruller will use GeoLite to determine City and Country names corresponding to client IP. It will determine the source IP by first looking at the "X-Forwarded-For" header. If not present, it will use the IP of the direct requestor.
* When Geolite is activated, the following attributes will be placed on input:
   * "\_ip\_country": Country name
   * "\_ip\_city": City name
   * "\_ip\_longitude": Longitude
   * "\_ip\_latitude: Latitude
   * "\_ip\_accuracy_radius: Accuracy radius

* When you pass a csv file in format "[country iso code],[City],[State]" using "--city-state-db", you will have an additional input:
   * "\_ip\_state: State based on city info

* You can define required inputs along with their associated types so that before processing rules Ruller will perform a basic check if they are present (ruller.AddRequiredInput(..)). This is usedful so that you don't have to perform those verifications inside each rule, as it was already verified before executing the rules.

## Request/Response filtering

* ```ruller.setRequestFilter(func(r *http.Request, input map[string]interface{}) error { return nil })```
   * You can verify http request attributes and change input map as you need

* ```func(w http.ResponseWriter, input map[string]interface{}, output map[string]interface{}, outBytes []byte) (bool, error) {return false, nil}```
   * You can verify the input and output map and write something to the response. If you return true, the default JSON marshal that ruller performs will be skipped.

* See an example at [sample/main.go](sample/main.go)

## More resources

* http://github.com/flaviostutz/ruller-sample-feature-flag - an example on how to build a DSL tool to generate Go Ruller code from a JSON and to build a Docker container with the REST api for your compiled rules. Has various functions for common scenarios of feature flags management

## Thanks, Maxmind!

This product includes GeoLite2 data created by MaxMind, available from
<a href="https://www.maxmind.com">https://www.maxmind.com</a>.
