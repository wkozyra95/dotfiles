{ buildGoModule }:
buildGoModule {
  name = "mycli";
  vendorHash = "sha256-EyGVG5ZpTVU43Y2bOlqgWUSRG1/00qAr/oBd87P3wyc=";
  src = ../..;
  subPackages = [ "." ];
  postFixup = ''
    cp $out/bin/dotfiles $out/bin/mycli
  '';
}

