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

  # Bundled mimalloc 2.0.9 uses ATOMIC_VAR_INIT, which was removed in C23
  # (GCC 15 default). Keep the C parts on C17.
  cmakeFlags = [ "-DCMAKE_C_STANDARD=17" ];
}
