let
  sources = import nix/sources.nix;
  pkgs = import sources.nixpkgs { };
  nix-npm-buildpackage = pkgs.callPackage sources.nix-npm-buildpackage {
    nodejs-10_x = pkgs.nodejs-13_x;
  };

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
    modSha256 = "1864h9gbyp4pmxwbxi28h0rn6sh120v0kyb3yviikdkkgjcw40my";
  };

  backend-image = pkgs.dockerTools.buildImage {
    name = "veilig-bellen-backend";
    tag = "latest";
    contents = pkgs.stdenvNoCC.initialPath
      ++ (with pkgs; [ backend bashInteractive busybox cacert ]);
    config.Cmd = [ "backend" ];
  };

in { inherit backend backend-image frontend-public; }
