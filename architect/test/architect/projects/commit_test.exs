defmodule Architect.Projects.CommitTest do
  use ExUnit.Case, async: true

  alias Architect.Projects.Commit
  alias Commit.Author

  doctest Commit

  describe "format/0" do
    test "correct format" do
      assert Commit.format() == "%H%n%aI%n%aE%n%aN%n%GF%n%s"
    end
  end

  describe "parse/1" do

    @output """
4c2439630bbea0bcad61adc78b434cc804117090
2019-01-17T17:39:27+00:00
vj@vjpatel.me
VJ Patel
%GF
add parsing of git shell for commits and branches
5697be45fa5cb5474a49f489c822e2d290693037
2019-01-16T23:21:41+00:00
vj@vjpatel.me
VJ Patel
%GF
Wip: added a few basic git funcs
"""


    test "Succesfully parses a commit" do

      {:ok, dt_first, _} = DateTime.from_iso8601("2019-01-17T17:39:27+00:00")
      {:ok, dt_second, _} = DateTime.from_iso8601("2019-01-16T23:21:41+00:00")

      expected = [
        %Commit{
          author: %Author{date: dt_first, email: "vj@vjpatel.me", name: "VJ Patel"},
          gpg_fingerprint: "%GF",
          message: "add parsing of git shell for commits and branches",
          sha: "4c2439630bbea0bcad61adc78b434cc804117090"
        },
        %Commit{
          author: %Author{date: dt_second, email: "vj@vjpatel.me", name: "VJ Patel"},
          gpg_fingerprint: "%GF",
          message: "Wip: added a few basic git funcs",
          sha: "5697be45fa5cb5474a49f489c822e2d290693037"
        }
      ]

      assert expected == Commit.parse(@output)

    end

  end
  describe "parse_show/1" do

    @output """
a51ab658cbb564bdf1990952b8eefb101c1aa823
2019-01-17T19:05:16+00:00
naedin@gmail.com
Eddy Lane
%GF
[architect] WIP
"""

    test "Succesfully parses a commit" do

      {:ok, dt, _} = DateTime.from_iso8601("2019-01-17T19:05:16+00:00")

      expected = %Commit{
        author: %Author{date: dt, email: "naedin@gmail.com", name: "Eddy Lane" },
        gpg_fingerprint: "%GF",
        message: "[architect] WIP",
        sha: "a51ab658cbb564bdf1990952b8eefb101c1aa823"
      }

      assert expected == Commit.parse_show(@output)

    end

  end

end
