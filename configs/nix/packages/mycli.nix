{ buildGoModule }:
buildGoModule {
  name = "mycli";
  vendorHash = "sha256-3y+1bkC9y9JiFl8qM6i9Gh42YR7RNBneZoJKGrWD6zs=";
  src = ../../..;
  subPackages = [ "." ];
  postFixup = ''
    cp $out/bin/dotfiles $out/bin/mycli
  '';
}

