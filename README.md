# IRMA veilig bellen

Consisting of a Go backend, a React agent website, and a button library.

## Development workflow

### To setup:

```bash
bin/setup.sh
```

### Database

After bringing up the docker containers for the first time, the database will
not have a schema yet. Use the following command to initialize it.

```bash
docker exec -i veilig-bellen_psql_1 psql -U tg -d tg < database/schema.sql
```

### To run:

```bash
bin/up.sh
```

### Manual flow

1. Go to `localhost:8080/session?purpose=foo`
2. Run `qrencode -t utf8` and paste the JSON
3. Scan the QR code with the Irma app and accept
4. When working with the old Irma app, it will not redirect you to the phone
   number. Look in the Irma server logs for the `clientReturnURL`. Go to
   `localhost:8080/call?dtmf=` followed by the last ten digits in the
   `clientReturnURL`.
5. Copy the secret it returns.
6. Go to `localhost:8080/disclose?secret=` followed by the copied secret. The
   Irma attributes should be returned.
