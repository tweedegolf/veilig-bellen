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

### Release artefacts

We use the Nix package manager to build our release artefacts. Nix will not
interfere with your system and will prevent your system's specifics from
interfering with this project's builds. Nix can be installed with:

    bash <(curl https://nixos.org/nix/install) --daemon

This will create a directory in root, `/nix`, which stores all Nix packages,
some nixbld users, a nix-daemon systemd service and make some changes to your
system's `bashrc` in order to extend your `PATH` with Nix. For more information,
see the [Nix manual](https://nixos.org/nix/manual/).

To build the public frontend locally:

    nix-build -A frontend-public

This will leave a symlink called `result` which points to the resulting frontend
library.

The build for the backend docker container is similar, but requires a `vendor`
directory for the backend as created by `go mod vendor`.

    nix-build -A backend-image

The resulting docker container can be loaded with:

    docker load < result

Or combine the steps with:

    docker load < $(nix-build -A backend-image --no-out-link)

Nix will cache builds and dependencies so builds are instant when no changes are
made. If `/nix` becomes too large, it can be cleaned up with
`nix-collect-garbage`.
