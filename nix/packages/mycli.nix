{ buildGoModule }:
buildGoModule {
  name = "mycli";
  vendorHash = "sha256-yBcMxUX9UsBd+qB3WkILcw++ug87SdHsNlMo2ywisaI=";
  src = ../..;
  subPackages = [ "." ];
  postFixup = ''
    cp $out/bin/dotfiles $out/bin/mycli
  '';
}

