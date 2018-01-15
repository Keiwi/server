# HTTP API documentation
This README describes the HTTP API routes for the server, such as updating group data, sending manual checks etc.

## Routes
 * /api/check
 * /api/command
 * /api/client
 * /api/group
 * /api/alert

## API
All API routes expects JSON data to be passed in the request body.

### Command
The command api handles 3 routes, insert, update and delete.

#### Insert
This route should be used when you add a group to a client

The expected JSON data looks like this:
```
{
    Type: String - insert/update/delete the API route,
    GroupName: String - The group name to insert into the cache service
    ID: The clients ID
} 
```

#### Update

### Client
### Group
### Alert
### Check

## Scenarios
This part will describe different scenarios when what API should be used