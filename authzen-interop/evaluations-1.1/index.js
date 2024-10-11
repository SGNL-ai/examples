const functions = require('@google-cloud/functions-framework');

const denyDecision = {
  decision: false,
  context: {
    reason: "Unsupported action"
  }
};

functions.http('evaluations', (req, res) => {
  const https = require("https");
  denyRequest = false;

  // Check if request has a content-type header indicating JSON
  if (!req.headers['content-type'] || !req.headers['content-type'].includes('application/json')) {
    console.error('Error: Invalid request. Body needs to be of type application/json');
    res.statusCode = 400;
    res.send('Error: Invalid request. Body needs to be of type application/json');
    return;
  }

  if (!req.headers['authorization'] || !req.headers['authorization'].startsWith('Bearer ')) { 
    console.error('Error: Authorization header not found');
    res.statusCode = 401;
    res.send('Error: Authorization header not found');
    return;
  }
  const bearerToken = req.headers['authorization'].split(' ')[1];

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
      "Authorization": "Bearer " + bearerToken
    }
  };

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
  if (action != 'can_update_todo') {
    denyRequest = true;
  }

  var assetIds = [];
  assetIds = getAssetIdsFromJsonBody(req.body);
  if (assetIds === null) {
    console.error('Error getting asset IDs from request body');
    res.send('Error getting asset IDs from request body');
    return;
  }

  if (action != 'can_update_todo') {
    console.error('Error: evaluations is only allowed on action can_update_todo for resources of type todos');
    res.send('Error: evaluations is only allowed on action can_update_todo for resources of type todos');
    return;
  } 

  const reqs = https.request(options, function (resp) {
    const chunks = [];

    resp.on("data", function (chunk) {
      chunks.push(chunk);
    });

    resp.on("end", function () {
      const body = Buffer.concat(chunks);
      const jsonBody = JSON.parse(body.toString());
      console.log(jsonBody);
      if (typeof jsonBody === 'object' && 'decisions' in jsonBody) {
        const decisions = jsonBody.decisions;
        var respDecisions = [];
        if (decisions.length > 0) {
          for (let i = 0; i < decisions.length; i++) {
            if ('decision' in decisions[i]) {
              if (decisions[i].decision == 'Allow') {
                respDecisions.push({decision: true});
              } else {
                respDecisions.push({decision: false});
              }
            } else {
              console.error('Error: "decision" field not found in decision');
              res.statusCode = 500;
              res.send('Error making request to SGNL Access API');
              return;
            }
          }
          res.statusCode = 200;
          res.send({"evaluations": respDecisions});
          return;
        } else {
          console.error('Error: No decisions found in response');
          res.statusCode = 500;
          res.send('Error making request to SGNL Access API');
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

  if (!denyRequest) {
    var queries = [];
    for (let i = 0; i < assetIds.length; i++) {
      queries.push({action: action, assetId: assetIds[i]});
    }
    var reqToSgnl = JSON.stringify({
      principal: {
        id: subject
      },
      queries: queries
    })
    console.log(reqToSgnl);
    reqs.write(reqToSgnl);
  } else {
    respDecisions = [];
    for (let i = 0; i < assetIds.length; i++) {
      respDecisions.push({decision: false});
    }
    res.statusCode = 200;
    res.send({"evaluations": respDecisions});
    return;
  }

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
      if ('id' in subject) {
        return subject.id;
      } else {
        console.error('Error getting subject from request body: "id" field not found');
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

function getAssetIdsFromJsonBody(body) {
  if (typeof body !== 'object' || !('evaluations' in body)) {
    console.error('Error getting asset IDs from evaluations: "evaluations" field not found');
    return null;
  }
  evaluations = body.evaluations;
  assetIds = [];
  for (let i = 0; i < evaluations.length; i++) {
    if ('resource' in evaluations[i] && 'properties' in evaluations[i].resource && 'ownerID' in evaluations[i].resource.properties) {
      if (!('id' in evaluations[i].resource)) {
      }
      assetIds.push(evaluations[i].resource.properties.ownerID);
    } else {
      console.error('Error getting asset IDs from evaluations: "resource" field not found');
      return null;
    }
  }
  return assetIds;
}