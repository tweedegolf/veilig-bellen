let
  sources = import nix/sources.nix;
  pkgs = import sources.nixpkgs { };
  nix-npm-buildpackage = pkgs.callPackage sources.nix-npm-buildpackage {
    nodejs-10_x = pkgs.nodejs-13_x;
  };

  inherit (nix-npm-buildpackage) buildYarnPackage;
  frontend-public = buildYarnPackage {
    name = "frontend-public.js";
    src = ./frontend-public;
    installPhase = ''
      mv dist/lib.js $out
    '';
  };

in {
  inherit frontend-public;
}
