{
  description = "Development Environment for GG Chatmix";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }: let
    system = "x86_64-linux";
    packageName = "gg-chatmix";
  in {
    devShells."${system}".default = let
      pkgs = import nixpkgs {
        inherit system;
      };
    in pkgs.mkShell {
      packages = with pkgs; [
        go
        libusb1
        pkg-config
        zsh
      ];

      shellHook = ''
        echo "`go version`"
        exec zsh
      '';
    };
  
    packages.${system} = let
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      gg-chatmix = pkgs.buildGoModule {
        name = packageName;
        src = pkgs.fetchFromGitHub {
          owner = "nilathedragon";
          repo = "gg-chatmix";
          rev = "e81b3cbe7325a7dd4af8c73347d1c2fbc1ae5a1b";
          sha256 = "sha256-V9Rg6qNRrS69I94hEaE3jAfzFLmkX8vHSR5EqDch5xY=";
        };
        vendorHash = "sha256-cmiC1BbBE9orgEkDpDesqP0Tf1kBW9WMIyRNtwxU1fU=";

        buildInputs = with pkgs; [
          libusb1
        ];

        nativeBuildInputs = with pkgs; [
          pkg-config
        ];
      };
    };

    defaultPackage.${system} = self.packages.${system}.gg-chatmix;

    nixosModule = let 
      pkgs = nixpkgs.legacyPackages.${system};
    in { config, lib, pkgs, ... }:
    {
      options.services.gg-chatmix = {
        enable = lib.mkEnableOption "enable the gg-chatmix service";
        package = lib.mkOption {
          type = lib.types.package;
          default = self.packages.${system}.gg-chatmix;
          description = "gg-chatmix package to use";
        };
      };

      config = lib.mkIf (config.services.gg-chatmix.enable) {
        systemd.user.services.gg-chatmix = {
          description = "GG Chatmix daemon";
          serviceConfig = {
            Type = "exec";
            ExecStart = "${config.services.gg-chatmix.package}/bin/gg-chatmix";
            Restart = "always";
          };
          wantedBy = [ "default.target" ];
        };

        services.udev.packages = [
          (pkgs.writeTextFile {
            name = "gg_chatmix_udev";
            text = ''SUBSYSTEMS=="usb", ATTRS{idVendor}=="1038", ATTRS{idProduct}=="12e0", TAG+="uaccess", MODE="0666"'';
            destination = "/etc/udev/rules.d/50-gg_chatmix.rules";
          })
        ];
      };
    };
  };
}