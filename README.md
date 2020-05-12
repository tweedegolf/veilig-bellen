# IRMA veilig bellen

IRMA veilig bellen is a project to allow municipalities to accept calls from
citizens authenticated with IRMA. This means the identity of the calling citizen
no longer needs to be checked by asking for their postal code and birth date.
Instead, the citizen will be guided to authenticate using the IRMA app, which
provides cryptographic proof of their credentials. The credentials requested can
be context-specific and can include their citizen identification number, their
address and their age. For more information on IRMA, see https://irma.app/

The project consists of a number of components:
- A backend which can be run multiple times to allow some amount of redundancy.
- A public frontend which guides citizens to make authenticated calls.
- An agent frontend which call center agents can use to accept calls and read
  any provided credentials.
- A panel frontend which shows metrics such as the number of available agents
  and the number of ongoing calls.
- An Amazon lambda used to connect Amazon Connect with the backends.
- A database schema to be used in a PostgreSQL database.

## Development workflow

The development workflow is based on docker-compose. You will need it and docker
installed. `docker-compose.yml` defines a cluster consisting of the three
frontends, any number of backends, a Postgres server, an IRMA server and an
Nginx server.

To set up a cluster of local docker containers, run:

```bash
bin/setup.sh
```

After bringing up the docker containers for the first time, the database will
not have a schema yet. Use the following command to initialize it.

```bash
docker exec -i veilig-bellen_psql_1 psql -U tg -d tg < database/schema.sql
```

To start the cluster with three backends:

```bash
bin/up.sh --scale backend=3
```

TODO: Explain how to set up Amazon Connect.

**Note:** in order to run with the metrics API enabled you need to create an Amazon IAM user, add the appropriate Amazon Connect roles,
and set `CONNECT_ID` and `CONNECT_SECRET` environment variables before starting the backend.

### Manual flow

TODO: Update to use the frontends and Amazon Connect.

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

## Manual build frontends

You can build your own frontend manually by running:

    docker-compose run -e BACKEND_URL="https://foo" frontend_public yarn run build-example

    docker-compose run -e BACKEND_URL="https://foo" -e CCP_HOST="example.awsapps.com" frontend_agents yarn run build

## Production

To bring the project into production, you will need to configure the following:

- `amazon/call-lambda.js` must be added as an Amazon lambda.
- Amazon Connect must be configured through a GUI. It must be configured to call
  that lambda when it receives a call with a DTMF code.
- A Postgres database must be configured and loaded with `database/schema.sql`.
- An IRMA server must be configured. See `docker-compose.yml` for example
  configuration.
- One or more backends must be configured. Travis CI provides up-to-date backend
  docker images for the master branch. Again, see `docker-compose.yml` for
  example configuration.
- TODO: frontends
- An Nginx server must be configured to serve the frontends, irma server and
  backend API over HTTPS. See `docker/nginx.conf` for an example configuration.

## Release artefacts

Our CI builds release artefacts for every commit in master. These can be
downloaded from Travis. TODO: link.

We use the Nix package manager to build our release artefacts. You will not need
it unless you need to make changes to CI. Nix will not interfere with your
system and will prevent your system's specifics from interfering with this
project's builds. Nix can be installed with:

    bash <(curl https://nixos.org/nix/install) --daemon

This will create a directory in root, `/nix`, which stores all Nix packages,
some nixbld users, a nix-daemon systemd service and make some changes to your
system's `bashrc` in order to extend your `PATH` with Nix. For more information,
see the Nix manual: https://nixos.org/nix/manual/

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

The backend image contains the backend as well as some basic unix tools. The
backend is located at `/bin/backend`.

Nix will cache builds and dependencies so builds are instant when no changes are
made. If `/nix` becomes too large, it can be cleaned up with
`nix-collect-garbage`.

To ensure reproducability, all dependencies are locked, both by the lockfiles of
the respective subprojects (`go.sum`, `yarn.lock`) and by `nix/sources.json`. To
update the lockfiles of the subprojects, see their respective tooling. To update
`nix/sources.json`, you will need `niv`. Install it with `nix-env -iA
nixpkgs.niv`. Run `niv update` from the repository root to update the
dependencies, then run the builds as before to check that they still build.
