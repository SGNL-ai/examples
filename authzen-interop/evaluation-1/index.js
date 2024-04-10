const functions = require('@google-cloud/functions-framework');

functions.http('helloHttp', (req, res) => {
  const https = require("https");

  const options = {
    "method": "POST",
    "hostname": "access.sgnlapis.cloud",
    "port": null,
    "path": "/access/v2/evaluations",
    "headers": {
      "Content-Type": "application/json",
      "Accept": "application/json",
      "X-Request-Id": "bfe9eb29-ab87-4ca3-be83-a1d5d8305716",
      "Accept-Language": "en, en-US;q=0.8, es;q=0.7",
      "Authorization": "Bearer " + process.env.authzen_token
    }
  };

  // Check if request has a content-type header indicating JSON
  if (!req.headers['content-type'] || !req.headers['content-type'].includes('application/json')) {
    return null;
  }
  
  subject = getSubjectFromJsonBody(req.body);
  if (subject === null) {
    console.error('Error getting subject from request body');
    res.send('Error getting subject from request body');
    return;
  }

  action = getActionFromJsonBody(req.body);
  if (action === null) {
    console.error('Error getting action from request body');
    res.send('Error getting action from request body');
    return;
  }

  resourceType = getResourceTypeFromJsonBody(req.body);
  if (resourceType === null) {
    console.error('Error getting resource type from request body');
    res.send('Error getting resource type from request body');
    return;
  }

  if (action == 'can_read_user' && resourceType != 'user') {
    console.error('Error: action can_read_user is only allowed for resources of type user');
    res.send('Error: action can_read_user is only allowed for resources of type user');
    return;
  } else if (action == 'can_read_user') {
    userId = getResourceUserIDFromJsonBody(req.body);
    if (userId == null) {
      console.error('Error: user ID is required for this action');
      res.send('Error: user ID is required for this action');
      return;
    }
    assetId = userId;
  } else {
    // action relates to "todo", and not user
    if (resourceType != 'todo') {
      console.error("Unknown resource type: " + resourceType);
      res.send("Unknown resource type: " + resourceType);
      return;
    } else {
      // resourceType is 'todo'
      resourceOwnerID = getResourceOwnerIDFromJsonBody(req.body);
      if ((action != 'can_create_todo' && action != 'can_read_todos') && resourceOwnerID == null) {
        console.error('Error: resource owner ID is required for this action');
        res.send('Error: resource owner ID is required for this action');
        return;
      } else {
        assetId = resourceOwnerID;
      }
    }
  }

  const reqs = https.request(options, function (resp) {
    const chunks = [];

    resp.on("data", function (chunk) {
      chunks.push(chunk);
    });

    resp.on("end", function () {
      const body = Buffer.concat(chunks);
      const jsonBody = JSON.parse(body.toString());
      if (typeof jsonBody === 'object' && 'decisions' in jsonBody) {
        const decisions = jsonBody.decisions;
        if (decisions.length > 0) {
          const decision = decisions[0];
          if ('decision' in decision) {
            if (decision.decision == 'Allow') {
              console.log('Access granted');
              res.statusCode = 200;
              res.send("{\"decision\": \"true\"}");
              return;
            } else {
              console.log('Access denied');
              res.statusCode = 200;
              res.send("{\"decision\": \"false\"}");
              return;
            }
          } else {
            console.error('Error: "decision" field not found in decision');
            res.statusCode = 500;
            res.send('Error: "decision" field not found in decision');
            return;
          }
        } else {
          console.error('Error: No decisions found in response');
          res.statusCode = 500;
          res.send('Error: No decisions found in response');
          return;
        }
      } else {
        console.error('Error: Decisions not found in response');
        console.log(body.toString());
        res.statusCode = 500;
        res.send('Error making request to SGNL Access API');
        return;
      }
    });
  });

  reqs.write(JSON.stringify({
    principal: {
      id: subject
    },
    queries: [{action: action, assetId: assetId}]
  }));
  reqs.on('error', function (error) {
    console.error('Error making request:', error.message);
    res.statusCode = 500;
    res.send('Error making request');
  });
  reqs.end();
});

function getSubjectFromJsonBody(body) {
  try {
    // Check if the parsed body is an object and has the "subject" field
    if (typeof body === 'object' && 'subject' in body) {
      const subject = body.subject;
      if ('identity' in subject) {
        return subject.identity;
      } else {
        console.error('Error getting subject from request body: "identity" field not found');
        return null;
      }
    } else {
      console.error('Error getting subject from request body: "subject" field not found');
      return null;
    }
  } catch (error) {
    // Handle potential parsing errors (invalid JSON)
    console.error('Error parsing request body:', error.message);
    return null;
  }
}

function getActionFromJsonBody(body) {
  try {
    // Check if the parsed body is an object and has the "action" field
    if (typeof body === 'object' && 'action' in body) {
      const action = body.action;
      if ('name' in action) {
        return action.name;
      } else {
        console.error('Error getting action from request body: "name" field not found');
        return null;
      }
    } else {
      console.error('Error getting action from request body: "action" field not found');
      return null;
    }
  } catch (error) {
    // Handle potential parsing errors (invalid JSON)
    console.error('Error parsing request body:', error.message);
    return null;
  }
}

function getResourceTypeFromJsonBody(body) {
  try {
    // Check if the parsed body is an object and has the "resource" field
    if (typeof body === 'object' && 'resource' in body) {
      const resource = body.resource;
      if ('type' in resource) {
        return resource.type;
      } else {
        console.error('Error getting resource from request body: "type" field not found');
        return null;
      }
    } else {
      console.error('Error getting resource from request body: "resource" field not found');
      return null;
    }
  } catch (error) {
    // Handle potential parsing errors (invalid JSON)
    console.error('Error parsing request body:', error.message);
    return null;
  }
}

function getResourceOwnerIDFromJsonBody(body) {
  try {
    // Check if the parsed body is an object and has the "resource" field
    if (typeof body === 'object' && 'resource' in body) {
      const resource = body.resource;
      if ('ownerID' in resource) {
        return resource.ownerID;
      } else {
        return null;
      }
    } else {
      return null;
    }
  } catch (error) {
    // Handle potential parsing errors (invalid JSON)
    console.error('Error parsing request body:', error.message);
    return null;
  }
}

function getResourceUserIDFromJsonBody(body) {
  try {
    // Check if the parsed body is an object and has the "resource" field
    if (typeof body === 'object' && 'resource' in body) {
      const resource = body.resource;
      if ('userID' in resource) {
        return resource.userID;
      } else {
        return null;
      }
    } else {
      return null;
    }
  } catch (error) {
    // Handle potential parsing errors (invalid JSON)
    console.error('Error parsing request body:', error.message);
    return null;
  }
}