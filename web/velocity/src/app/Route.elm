module Route exposing (Route(..), href, modifyUrl, fromLocation)

import UrlParser as Url exposing (parseHash, s, (</>), string, oneOf, Parser)
import Navigation exposing (Location)
import Html exposing (Attribute)
import Html.Attributes as Attr
import Data.Project as Project
import Data.Commit as Commit
import Page.Project.Route as ProjectRoute exposing (Route(..))


type Route
    = Home
    | Login
    | Logout
    | Projects
    | Project ProjectRoute.Route Project.Id
    | KnownHosts


route : Parser (Route -> a) a
route =
    oneOf
        [ Url.map Home (s "")
        , Url.map Login (s "sign-in")
        , Url.map Logout (s "logout")
        , Url.map Projects (s "projects")
        , Url.map KnownHosts (s "known-hosts")
        , Url.map (Project Commits) (s "project" </> Project.idParser </> s "commits")
        , Url.map (\id hash -> Project (Commit hash) id) (s "project" </> Project.idParser </> s "commits" </> Commit.hashParser)
        , Url.map (Project Settings) (s "project" </> Project.idParser </> s "settings")
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

                Project child id ->
                    [ "project", Project.idToString id ] ++ (ProjectRoute.routeToPieces child)

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
