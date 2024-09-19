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

          caddy-log-exporter = pkgs.buildGo123Module {
            name = "caddy-log-exporter";
            src = ./.;
            vendorHash = null;
            ldflags = [ "-s" "-w" "-extldflags '-static'" ];
            CGO_ENABLED = 0;
          };

          docker = pkgs.dockerTools.buildImage {
            name = "xsteadfastx/caddy-log-exporter";
            tag = "latest";
            config = {
              Entrypoint = [ "${caddy-log-exporter}/bin/caddy-log-exporter" ];
            };
          };

          env-pkgs = with pkgs; [
            go_1_23

            go-task
            gofumpt
            golangci-lint
            golines
            gopls

            skopeo
          ];
        in
        with pkgs; {
          packages =
            {
              default = caddy-log-exporter;
              inherit caddy-log-exporter docker;
            };

          devShells.default = mkShell {
            buildInputs = env-pkgs;
          };
        }
      );
}

