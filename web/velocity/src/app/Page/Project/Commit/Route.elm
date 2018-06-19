module Page.Project.Commit.Route exposing (Route(..), routeToPieces, route)

import UrlParser as Url exposing (parseHash, s, (</>), (<?>), string, stringParam, intParam, oneOf, Parser)
import Data.Task as Task
import Data.Build as Build
import Data.Commit as Commit
import Util exposing ((=>))


type Route
    = Overview
    | Task Task.Name (Maybe Build.Id)


route : Parser (Route -> b) b
route =
    oneOf
        [ Url.map Overview (s "overview")
        , Url.map Task (s "tasks" </> Task.nameParser <?> (Build.idQueryParser "build"))
        ]



-- PUBLIC HELPERS --


routeToPieces : Route -> ( List String, List ( String, String ) )
routeToPieces page =
    case page of
        Overview ->
            [ "overview" ] => []

        Task name maybeBuildId ->
            let
                queryParams =
                    maybeBuildId
                        |> Maybe.map (\id -> [ "build" => Build.idToString id ])
                        |> Maybe.withDefault []
            in
                [ "tasks", Task.nameToString name ] => queryParams
