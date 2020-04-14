// TODO change to https
const http = require('http');
const querystring = require('querystring');

// Change this to your own IRMA Veilig Bellen backend URL.
const URL = "http://proxy.irma.bellen.tweede.golf/call";

exports.handler = (data, context, callback) => {
    const attributes = data.Details.ContactData.Attributes;
    console.log('attributes', attributes);
    const req = http.request(URL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
    }, (res) => {
        let body = '';
        console.log('Status:', res.statusCode);
        console.log('Headers:', JSON.stringify(res.headers));
        res.setEncoding('utf8');
        res.on('data', (chunk) => body += chunk);
        res.on('end', () => {
            if (res.statusCode !== 200) {
                callback(Error(`Backend ${res.statusCode}`));
            } else {
                const results = { session_secret: body };
                console.log('results', results);
                callback(null, results);
            }
        });
    });
    req.on('error', callback);
    req.write(querystring.stringify(attributes));
    req.end();
};
