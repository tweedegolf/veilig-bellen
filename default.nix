let
  sources = import nix/sources.nix;
  pkgs = import sources.nixpkgs { };
  nix-npm-buildpackage = pkgs.callPackage sources.nix-npm-buildpackage {
    nodejs-10_x = pkgs.nodejs-13_x;
  };

  vgo2nix = import sources.vgo2nix { inherit pkgs; };

  inherit (nix-npm-buildpackage) buildYarnPackage;
  inherit (pkgs.nix-gitignore) gitignoreSource;

  frontend-public = buildYarnPackage {
    name = "veilig-bellen-frontend-public.js";
    src = gitignoreSource [ ] ./frontend-public;
    installPhase = ''
      mv dist/lib.js $out
    '';
  };

  backend = pkgs.buildGoModule {
    name = "veilig-bellen-backend";
    src = gitignoreSource [ ] ./backend;
    modSha256 = "";
    configurePhase = ''
      runHook preConfigure
      export GOCACHE=$TMPDIR/go-cache
      export GOPATH="$TMPDIR/go"
      export GOSUMDB=off
      ln -s ${./backend/vendor} vendor
      cd "$modRoot"
      runHook postConfigure
    '';
  };

  backend-image = pkgs.dockerTools.buildImage {
    name = "veilig-bellen-backend";
    tag = "latest";
    contents = pkgs.stdenvNoCC.initialPath
      ++ (with pkgs; [ backend bashInteractive busybox cacert ]);
    config.Cmd = [ "backend" ];
  };

in { inherit backend backend-image frontend-public; }
