# Web app for Digital Matters Yabby3 tracking devices

This project is an HTTP server that provides an endpoint for the [Digital
Matters Yabby3](https://support.digitalmatter.com/en_US/yabby3-cellular) GPS
tracking devices to send data to. It also serves web pages of maps with pins
showing current tag locations.

[<img src="screenshot.png">]()

Notifications indicate when the tag has left boundaries.

[<img src="ntfy-screenshot.png">]()


## Theory of operation

A service written in Go listens to endpoints.

`/upload` is for devices to post their JSON payloads The payload is
dumped as-is into a MongoDB database.

If the tag is outside the Safe Zone or Property boundaries, notifications are
sent (currently only [ntfy.sh](https://ntfy.sh) is supported).

A boundary can be divided up into named zones for more useful notifications.
These zones are defined in .kml files created with Google Earth.

When web users visit `/current` with a browser, a Google Map is returned, with
markers showing current tag positions. The markers have other data about the
tag, including data freshness and battery level.


## Installation and setup

1. You'll need a VPS and a Google Maps API key. The VPS needs to have MongoDB
   installed.
2. Notifications are currently supported using the incredibly easy
   [ntfy.sh](https://ntfy.sh). No signup or API key is required to use it. Just
   install the app on your phone and choose a random subscription name.
3. Customize environment variables in devops/dog-tracking.service
4. Set your VPS name in the deployment script in devops/deploy_and_run.sh
5. Delete the example zone and boundary files from `boundary_zones/` and `named_zones/`.
6. Generate .kml zone and boundary files using Google Earth and save those files
   individually to `boundary_zones/` and `named_zones/`.
7. Create an issue to bug the author to make the rest of the system configurable
   (see TODOs below)



## Deployment

From this dir, run

    $ devops/deploy_and_run.sh


## To run tests

    $ docker run --name some-mongo -d mongo:7.0
    $ go test ./...


## TODOs

1. Make own location configurable (get it out of current-map.html).
2. Read boundaries from boundary_zones dir (get them out of main.go).
3. Put all boundaries and zones into two .kml files (just save the top folder in Google Earth).
4. Add instructions about how to make zones in Google Earth.
5. Improve instructions for setting up a server.
6. Make domain configurable with env vars (get it out of ntfy.go).
7. Make dog names and tag IDs controlled by env vars.
