{ stdenv, fetchFromGitHub, cmake }:

stdenv.mkDerivation {
  pname = "lua-code-format";
  version = "1.4.2";

  src = fetchFromGitHub {
    owner = "CppCXY";
    repo = "EmmyLuaCodeStyle";
    rev = "1b5763ce26b7112972e83f84ec140941497575f8";
    hash = "sha256-bZFMk2vRIVJu5LzpVKC9ZAjW9wNRdnf1KEBITktpaAY=";
    fetchSubmodules = true;
  };

  nativeBuildInputs = [
    cmake
  ];
}
