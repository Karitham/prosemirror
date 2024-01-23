{
  description = "dev shell for prosemirror development";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils/main";
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      rec {
        devShell = pkgs.mkShell {
          name = "digtate";
          packages = with pkgs; [
            go_1_22
            gofumpt
            delve

            # to run the examples without fuss
            bun
          ];

          # delve buggy if hardening set
          hardeningDisable = [ "fortify" ];
        };
      });
}
