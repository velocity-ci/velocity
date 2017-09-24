module Page.Project.Route exposing (Route(..), routeToPieces)

import UrlParser as Url exposing (parseHash, s, (</>), string, oneOf, Parser)
import Navigation exposing (Location)
import Html.Attributes as Attr
import Data.Project as Project exposing (Project)
import Data.Commit as Commit


type Route
    = Commits
    | Commit Commit.Hash
    | Settings



-- PUBLIC HELPERS --


routeToPieces : Route -> List String
routeToPieces page =
    case page of
        Commits ->
            [ "commits" ]

        Commit hash ->
            [ "commits", Commit.hashToString hash ]

        Settings ->
            [ "settings" ]
