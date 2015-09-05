# Jarvis Bot

A Telegram bot, built over 4 days, for friends. Current featureset includes: 

* Grabs images
* Grabs GIFs
* Does Google searches
* Does exchange rate conversions
* Displays the current air pollution index for all areas in Singapore
* Clears your NSFW stuff

## Build dependencies
Jarvis relies on go-bindata to package assets in the data/ folder into the
binary. If you'd like to add to the assets when implementing your own function, 
install the go-bindata tool using:

```go get -u github.com/jteeuwen/go-bindata/... ```

If you're running a Go version < 1.4, you'll need to manually run the following
command in the top level dir. (Otherwise, you can run `go generate` to achieve
the same results).

```go-bindata -pkg jarvisbot -o assets.go data/```

## Instructions 
Compile Jarvis for your target platform and upload the binary to your server. 

In the same directory as the binary, include a `temp` directory, and a config.json file. (A
config-sample.json has been provided for you.) Remember to include the API keys
in your config.json! 

Last, but not least, make sure the user running the program has write permissions to its directory, 
as Jarvis creates a jarvis.db file.
