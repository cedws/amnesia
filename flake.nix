{
  description = "amnesia";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        buildGoModule = if pkgs ? buildGo125Module then pkgs.buildGo125Module else pkgs.buildGo124Module;
      in
      {
        packages.amnesia = buildGoModule {
          pname = "amnesia";
          version = "0.0.0";
          src = ./.;
          subPackages = [ "." ];
          vendorHash = "sha256-QgnkvL+GVRM5vnzgVk+C3PJW4onNt3p0yeaH+pdtfvA=";
          env = {
            CGO_ENABLED = "0";
          };
          ldflags = [
            "-s"
            "-w"
          ];
          meta = with pkgs.lib; {
            description = "amnesia is a command-line tool for sealing secrets with a set of questions";
            homepage = "https://github.com/cedws/amnesia";
            license = licenses.gpl3Only;
            mainProgram = "amnesia";
            platforms = platforms.unix;
          };
        };

        packages.default = self.packages.${system}.amnesia;
      }
    );
}
