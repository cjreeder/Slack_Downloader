# Slack Downloader Script

This useful little script/app can be run on any channel in a slack instance that has the following bot API Privileges - 
* channels:history
* channels:read
* files:read
* groups:history
* groups:read
* im:history
* im:read
* mpim:history
* mpim:read
* users:read

Parameters to pass to the script:
* bearer token for bot with above privileges
* channel in Slack Instance that bot has been added to
* local location where all of this is stored

Script will output the whole history of the channel to file in JSON format and all pictures will be dropped into the same directory as well with time stamps from when they were uploaded.  
