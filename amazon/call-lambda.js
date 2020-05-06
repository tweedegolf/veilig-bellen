// Set CALL_URL to something like http://proxy.irma.bellen.tweede.golf/call

const http = require('http');
const querystring = require('querystring');

exports.handler = (data, _context, callback) => {
    const attributes = data.Details.ContactData.Attributes;
    const url = process.env.CALL_URL;
    const req = http.request(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    }, (res) => {
        let body = '';
        res.setEncoding('utf8');
        res.on('data', (chunk) => body += chunk);
        res.on('end', () => {
            if (res.statusCode !== 200) {
                callback(Error(`Backend ${res.statusCode}`));
            } else {
                const results = { session_secret: body };
                callback(null, results);
            }
        });
    });
    req.on('error', callback);
    req.write(querystring.stringify(attributes));
    req.end();
};
