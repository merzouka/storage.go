# About
this is a back-end written in __Go__, that uploads files and stores them in different locations (writer); basically a distributed storage.  
The app includes:
- a health check to verify which locations are available
- a rollback feature in case any of the locations fail to fulfill the request
- the ability to add meta-data to uploaded files
- the ability to have multiple versions of the same file and get each one
The app is meant to be run on kubernetes and the configuration is in the __config__ folder.  
API requests are handled by a reverse proxy (proxy).   
The app comes with a simple ui to interact with the API

# Running
- install [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- run `kubectl apply -f config/` to get the app running
- run `kubectl port-forward svc/proxy 8080:8080` to expose the API
- open the UI in a browser window
- upload as many files as you want
Some sample files are included in the `test` folder
