let
  sources = import nix/sources.nix;
  pkgs = import sources.nixpkgs { };
  vgo2nix = import sources.vgo2nix { inherit pkgs; };

in pkgs.mkShell { buildInputs = [ vgo2nix ]; }
