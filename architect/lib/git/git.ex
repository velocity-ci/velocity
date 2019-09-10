defmodule Git do
  @spec version() :: binary()
  def version() do
    {out, 0} = System.cmd("git", ["--version"])
    String.slice(String.trim(out), 12..-1)
  end
end
