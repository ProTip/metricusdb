#metricusdb
metricusdb is an experimental time series database using Riak as the default backend.

##Goals
* Efficiently store ephemeral time series
* Optional sub-second resolution
* Variable rate
* Support for backfilling and delayed metrics
* Qeury subscriptions for receiving new datapoints via a push mechanism(Websockets, etc)
* Queries as pre-processing pipelines
* Ability to query across in-cache and stored metrics
* Index streams for easy searching(Riak Search 2.0)
* Graphite API compatiblity to leverage existing tools ecosystem(provided via an API lexar/parser and comptability layer)
