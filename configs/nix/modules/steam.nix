username:
{ pkgs, ...}:
{
  programs.steam = {
    enable = true;
  };
  users.users.${username} = {
	  packages = with pkgs; [
      steam
      steam-run
    ];
  };
}
