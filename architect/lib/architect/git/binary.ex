defmodule Architect.Git.Binary do
  def version() do
    res = Porcelain.exec("git", ["--version"])
    String.slice(String.trim(res.out), 12..-1)
  end
end
