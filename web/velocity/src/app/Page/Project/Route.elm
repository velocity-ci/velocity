module Page.Project.Route exposing (Route(..), routeToPieces)

import UrlParser as Url exposing (parseHash, s, (</>), string, oneOf, Parser)
import Navigation exposing (Location)
import Html exposing (Attribute)
import Html.Attributes as Attr
import Data.Project as Project


type Route
    = Commits
    | Settings



-- PUBLIC HELPERS --


routeToPieces : Route -> List String
routeToPieces page =
    case page of
        Commits ->
            [ "commits" ]

        Settings ->
            [ "settings" ]
