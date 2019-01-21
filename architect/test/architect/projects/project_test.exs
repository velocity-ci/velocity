defmodule Architect.Projects.ProjectTest do
  use ExUnit.Case, async: true

  alias Architect.Projects.Project

  doctest Project

  describe "project" do
    test "name_from_http_address/1 parses correctly" do
      mappings = %{
        "https://github.com/foo/bar.git": "foo/bar @ github.com",
        "https://gitlab.com/foo/bar.git": "foo/bar @ gitlab.com"
      }

      Enum.each(mappings, fn {address, expected} ->
        name = Project.name_from_http_address(Atom.to_string(address))
        assert name == expected
      end)
    end

    test "name_from_git_address/1 parses correctly" do
      mappings = %{
        "git@github.com:foo/bar.git": "foo/bar @ github.com",
        "git@gitlab.com:foo/bar.git": "foo/bar @ gitlab.com"
      }

      Enum.each(mappings, fn {address, expected} ->
        name = Project.name_from_git_address(Atom.to_string(address))
        assert name == expected
      end)
    end
  end
end
