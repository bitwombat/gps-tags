# Web app for Digital Matters Yabby3 tracking devices
This service provides an endpoint for the Digital Matters Yabby3 tracking
devices to send data to, as well as providing web pages of maps and pins showing
current tag locations.

## Theory of operation

Devices are configured to post their JSON payloads to `/upload`. The payload is
dumped as-is into a MongoDB database.

If the tag is outside the Safe Zone or Propety boundaries, notifications are
sent (see below).

When users visit `/current`, the database is queried, and users are presented with a map as shown below.

## Installation/setup

1. You'll need a VPS and a Google Maps API key. The VPS should have MongoDB
   installed.
2. Notifications are currently supported using [ntfh.sh](https://ntfy.sh). No signup or API key is required to use this service. Just choose a random subscription name.
3. Customize environment variables in devops/dog-tracking.service
4. Customize the deployment script in devops/deploy_and_run.sh
5. Delete the example zone and boundary files from `boundary_zones/` and `named_zones/`.
6. Generate .kml zone and boundary files using Google Earth and save those files
   individually to `boundary_zones/` and `named_zones/`.
7. Create an issue to bug the author to make the rest of the system configurable
   (tag IDs, animal names, service URL).

## Deployment

From this dir, run

    $ devops/deploy_and_run.sh

## To run tests

    $ docker run --name some-mongo -d mongo:7.0
    $ go test ./...
