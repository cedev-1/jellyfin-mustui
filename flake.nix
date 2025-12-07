{
  description = "Flake template";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      forAllSystems = function:
        nixpkgs.lib.genAttrs [
          "x86_64-linux"
          "aarch64-linux"
          "x86_64-darwin"
          "aarch64-darwin"
        ] (system: function nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "jellyfin-mustui";
          version = "0.1.0";
          src = ./.;
          
          vendorHash = "sha256-2Pbg9EWFoq9O5yuDheUEiJSQK7UrKzNwSxXbzZUDeII=";

          nativeBuildInputs = [ pkgs.pkg-config ];

          buildInputs = [ pkgs.alsa-lib ];

          meta = {
            description = "A TUI music player for Jellyfin in Go";
            homepage = "https://github.com/cedev-1/jellyfin-mustui";
            mainProgram = "jellyfin-mustui";
          };
        };
      });

      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            pkg-config
          ];
          buildInputs = [
            pkgs.alsa-lib
          ];
        };
      });
    };
}
