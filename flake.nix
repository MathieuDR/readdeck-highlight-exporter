{
  description = "Go flake for my Readdeck highlight exporter";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    systems,
    nixpkgs,
    treefmt-nix,
    ...
  }: let
    inherit (nixpkgs) lib;
    eachSystem = f: lib.genAttrs (import systems) (system: f nixpkgs.legacyPackages.${system});
    treefmtEval = eachSystem (pkgs: treefmt-nix.lib.evalModule pkgs ./treefmt.nix);
    version = builtins.substring 0 8 (self.lastModifiedDate or "19700101");
    exe_name = "highlight-exporter";
  in {
    # Build executables. See https://nixos.org/manual/nixpkgs/stable/#sec-language-go
    packages = eachSystem (pkgs: {
      default = pkgs.buildGoModule {
        pname = exe_name;
        version = version;
        src = self.outPath;
        vendorHash = "sha256-+V4LaP3PP8KdcEx1iFWrhVQ1YjXqI51vzhcqA6VI36k=";
        meta = {};

        # Ensure the binary is named correctly
        postInstall = ''
          mv $out/bin/* $out/bin/${exe_name}
        '';

        preBuild = ''
          echo "Building version: ${version}"
          echo "Using ldflags: -X main.ProgramName=${exe_name} -X main.Version=${version} -X main.BuildTime=nixbuild"
        '';

        ldflags = [
          "-X github.com/mathieudr/readdeck-highlight-exporter/cmd.programName=${exe_name}"
          "-X github.com/mathieudr/readdeck-highlight-exporter/cmd.version=${version}"
          "-X github.com/mathieudr/readdeck-highlight-exporter/cmd.buildTime=nixbuild"
        ];

        meta = {
          description = "A readdeck highlight exporter";
          homepage = "https://github.com/MathieuDR/readdeck-highlight-exporter";
          license = lib.licenses.mit;
          maintainers = ["MathieuDR"];
        };
      };
    });

    devShells = eachSystem (pkgs: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          go-tools
          gotools
          delve
        ];

        shellHook = ''
          echo "Welcome to the readdeck highlight exporter development environment!"
          echo "Go version: $(go version)"

          mkdir -p $PWD/.nix/go/bin
          export PATH=$PWD/.nix/go/bin:$PATH
          export GOBIN=$PWD/.nix/go/bin

          echo "Project binaries will be installed to: $GOBIN"
          echo ""

          echo "Available commands:"
          echo "  go build              - Build the project"
          echo "  go test               - Run tests"
          echo "  go run main.go        - Run the project"
          echo "  golangci-lint run     - Run linters"
          echo "  dlv debug             - Debug the application"
        '';
      };
    });

    formatter = eachSystem (pkgs: treefmtEval.${pkgs.system}.config.build.wrapper);

    checks = eachSystem (pkgs: {
      formatting = treefmtEval.${pkgs.system}.config.build.check self;
      build = self.packages.${pkgs.system}.default;
    });
  };
}
