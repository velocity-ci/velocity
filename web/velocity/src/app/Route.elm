module Route exposing (Route(..), href, modifyUrl, fromLocation, routeToString)

import UrlParser as Url exposing (parsePath, s, (</>), string, oneOf, Parser)
import Navigation exposing (Location)
import Html exposing (Attribute)
import Html.Attributes as Attr
import Data.Project as Project
import Page.Project.Route as ProjectRoute
import Util exposing ((=>))


type Route
    = Home
    | Login
    | Logout
    | Projects
    | Users
    | Project Project.Slug ProjectRoute.Route
    | KnownHosts


route : Parser (Route -> a) a
route =
    oneOf
        [ Url.map Home (s "")
        , Url.map Login (s "sign-in")
        , Url.map Logout (s "logout")
        , Url.map Projects (s "projects")
        , Url.map KnownHosts (s "known-hosts")
        , Url.map Users (s "users")
        , Url.map (\slug -> Project slug ProjectRoute.default) (s "projects" </> Project.slugParser)
        , Url.map Project (s "projects" </> Project.slugParser </> ProjectRoute.route)
        ]



-- INTERNAL --


routePieces : Route -> ( List String, List ( String, String ) )
routePieces page =
    case page of
        Home ->
            [] => []

        Login ->
            [ "sign-in" ] => []

        Logout ->
            [ "logout" ] => []

        Projects ->
            [ "projects" ] => []

        Users ->
            [ "users" ] => []

        Project slug child ->
            let
                ( subPath, subQuery ) =
                    ProjectRoute.routeToPieces child
            in
                ( [ "projects", Project.slugToString slug ] ++ subPath, subQuery )

        KnownHosts ->
            [ "known-hosts" ] => []


routeToString : Route -> String
routeToString page =
    let
        pieces =
            routePieces page

        path =
            Tuple.first pieces
                |> String.join "/"

        queryString =
            Tuple.second pieces
                |> List.map (\( k, v ) -> k ++ "=" ++ v)
                |> String.join "&"

        routeString =
            "/" ++ path
    in
        if String.length queryString > 0 then
            routeString ++ "?" ++ queryString
        else
            routeString



-- PUBLIC HELPERS --


href : Route -> Attribute msg
href route =
    Attr.href (routeToString route)


modifyUrl : Route -> Cmd msg
modifyUrl =
    routeToString >> Navigation.modifyUrl


fromLocation : Location -> Maybe Route
fromLocation location =
    if String.isEmpty location.pathname then
        Just Home
    else
        parsePath route location
