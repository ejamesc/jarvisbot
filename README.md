<h3 align="center">
  <div align="center">
    <h1>Jarvis</h1>
    <h6>A Telegram bot for friends</h6>
  </div>
  <a href="https://github.com/ejamesc/jarvisbot">
    <img src="http://i.imgur.com/yZBaf9T.png" alt="Jarvis" width="500" />
  </a>
</h3>

------

Current featureset includes: 

* Grabs images
* Grabs GIFs
* Does Google searches
* Does Youtube searches
* Does location searches
* Does exchange rate conversions
* Displays the current air pollution index for all areas in Singapore
* Clears your NSFW stuff

## Build dependencies
Jarvis relies on go-bindata to package assets in the data/ folder into the
binary. If you'd like to add to the assets when implementing your own response functions, 
install the go-bindata tool using:

```go get -u github.com/jteeuwen/go-bindata/... ```

If you're running a Go version < 1.4, you'll need to manually run the following
command in the top level dir. (Otherwise, you can run `go generate` to achieve
the same results).

```go-bindata -pkg jarvisbot -o assets.go data/```

## Instructions 
Compile Jarvis for your target platform and upload the binary to your server. 

In the same directory as the binary, create a config.json file. (A
config-sample.json has been provided for you to modify.) Remember to include the API keys
in your config.json! 

Last, but not least, make sure the user running the program has write permissions to its directory, 
as Jarvis creates a jarvis.db file.

## Screenshots

<img src="http://i.imgur.com/GldfYIX.png" alt="Jarvis" width="500" />
<img src="http://i.imgur.com/uGmOQIC.png" alt="Jarvis" width="500" />
<img src="http://i.imgur.com/EWMJEoF.png" alt="Jarvis" width="500" />
