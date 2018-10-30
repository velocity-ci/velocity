module Page.Project.Route exposing (Route(..), default, route, routeToPieces)

import Data.Branch as Branch
import Data.Commit as Commit
import Page.Project.Commit.Route as CommitRoute
import UrlParser as Url exposing ((</>), (<?>), Parser, intParam, oneOf, parseHash, s, string, stringParam)
import Util exposing ((=>))


type Route
    = Overview
    | Builds (Maybe Int)
    | Commits (Maybe Branch.Name) (Maybe Int)
    | Commit Commit.Hash CommitRoute.Route
    | Settings


default : Route
default =
    Commits Nothing Nothing


route : Parser (Route -> b) b
route =
    oneOf
        [ Url.map Overview (s "overview")
        , Url.map Settings (s "settings")
        , Url.map Builds (s "builds" <?> intParam "page")
        , Url.map Commits (s "commits" </> Branch.nameParser <?> intParam "page")
        , Url.map Commit (s "commit" </> Commit.hashParser </> CommitRoute.route)
        ]



-- PUBLIC HELPERS --


routeToPieces : Route -> ( List String, List ( String, String ) )
routeToPieces page =
    case page of
        Overview ->
            [] => []

        Builds maybePageNumber ->
            let
                queryParams =
                    case maybePageNumber of
                        Just 1 ->
                            []

                        Just p ->
                            [ ( "page", toString p ) ]

                        _ ->
                            []
            in
                [ "builds" ] => queryParams

        Commits branchName maybePageNumber ->
            let
                queryParams =
                    case maybePageNumber of
                        Just 1 ->
                            []

                        Just p ->
                            [ ( "page", toString p ) ]

                        _ ->
                            []
            in
                [ "commits", Branch.nameToString branchName ] => queryParams

        Commit hash child ->
            let
                ( subPath, subQuery ) =
                    CommitRoute.routeToPieces child
            in
                [ "commit", Commit.hashToString hash ] ++ subPath => subQuery

        Settings ->
            [ "settings" ] => []
