module Route exposing (Route(..), href, modifyUrl, fromLocation)

import UrlParser as Url exposing (parseHash, s, (</>), string, oneOf, Parser)
import Navigation exposing (Location)
import Html exposing (Attribute)
import Html.Attributes as Attr
import Data.Project as Project
import Page.Project.Route as ProjectRoute


type Route
    = Home
    | Login
    | Logout
    | Projects
    | ProjectChild ProjectRoute.Route Project.Id
    | KnownHosts


route : Parser (Route -> a) a
route =
    oneOf
        [ Url.map Home (s "")
        , Url.map Login (s "sign-in")
        , Url.map Logout (s "logout")
        , Url.map Projects (s "projects")
        , Url.map KnownHosts (s "known-hosts")
        , Url.map (\id -> ProjectChild (ProjectRoute.Commits id) id) (s "project" </> Project.idParser </> s "commits")
        ]



-- INTERNAL --


routeToString : Route -> String
routeToString page =
    let
        pieces =
            case page of
                Home ->
                    []

                Login ->
                    [ "sign-in" ]

                Logout ->
                    [ "logout" ]

                Projects ->
                    [ "projects" ]

                ProjectChild child id ->
                    [ "project", Project.idToString id, "commits" ]

                KnownHosts ->
                    [ "known-hosts" ]
    in
        "#/" ++ (String.join "/" pieces)



-- PUBLIC HELPERS --


href : Route -> Attribute msg
href route =
    Attr.href (routeToString route)


modifyUrl : Route -> Cmd msg
modifyUrl =
    routeToString >> Navigation.modifyUrl


fromLocation : Location -> Maybe Route
fromLocation location =
    if String.isEmpty location.hash then
        Just Home
    else
        parseHash route location
