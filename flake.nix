{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs:
    inputs.flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ ];
          };

          env-pkgs = with pkgs; [
            go_1_23

            go-task
            gofumpt
            golangci-lint
            golines
            gopls
            goreleaser
          ];
        in
        with pkgs; {
          devShells.default = mkShell {
            buildInputs = env-pkgs;
          };
        }
      );
}

