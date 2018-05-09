module Views.Build exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStreamOutput)
import Data.Task as ProjectTask
import Ansi.Log
import Util exposing ((=>))


viewBuildContainer : Build -> List BuildStep -> Ansi.Log.Model -> Html msg
viewBuildContainer build steps output =
    div []
        [ h3 [] [ text ("build container " ++ Build.idToString build.id) ]
        , Ansi.Log.view output
        ]


viewBuildStatusIcon : Build -> Html msg
viewBuildStatusIcon build =
    case build.status of
        Build.Waiting ->
            i [ class "fa fa-clock-o" ] []

        Build.Running ->
            i [ class "fa fa-cog fa-spin" ] []

        Build.Success ->
            i [ class "fa fa-check" ] []

        Build.Failed ->
            i [ class "fa fa-times" ] []


viewBuildTextClass : Build -> String
viewBuildTextClass build =
    case build.status of
        Build.Waiting ->
            "text-secondary"

        Build.Running ->
            "text-primary"

        Build.Success ->
            "text-success"

        Build.Failed ->
            "text-danger"


viewBuildStepStatusIcon : BuildStep -> Html msg
viewBuildStepStatusIcon buildStep =
    case buildStep.status of
        BuildStep.Waiting ->
            i [ class "fa fa-clock-o" ] []

        BuildStep.Running ->
            i [ class "fa fa-cog fa-spin" ] []

        BuildStep.Success ->
            i [ class "fa fa-check" ] []

        BuildStep.Failed ->
            i [ class "fa fa-times" ] []


streamBadgeClass : Int -> String
streamBadgeClass index =
    case index of
        0 ->
            "badge-primary"

        1 ->
            "badge-secondary"

        2 ->
            "badge-success"

        3 ->
            "badge-danger"

        4 ->
            "badge-warning"

        5 ->
            "badge-info"

        _ ->
            "badge-dark"


headerBackgroundColourClassList : BuildStep -> List ( String, Bool )
headerBackgroundColourClassList { status } =
    case status of
        BuildStep.Waiting ->
            []

        BuildStep.Running ->
            []

        BuildStep.Success ->
            [ "text-success" => True
            , "bg-transparent" => True
            ]

        BuildStep.Failed ->
            [ "bg-transparent" => True
            , "text-danger" => True
            ]


buildStepBorderColourClassList : BuildStep -> List ( String, Bool )
buildStepBorderColourClassList { status } =
    case status of
        BuildStep.Waiting ->
            [ "border" => True
            ]

        BuildStep.Running ->
            [ "border" => True
            , "border-primary" => True
            ]

        BuildStep.Success ->
            [ "border" => True
            ]

        BuildStep.Failed ->
            [ "border" => True
            ]


buildCardClassList : Build -> List ( String, Bool )
buildCardClassList { status } =
    case status of
        Build.Success ->
            [ "text-success" => True
            ]

        Build.Failed ->
            [ "text-danger" => True
            ]

        _ ->
            []


viewCardTitle : ProjectTask.Step -> String
viewCardTitle taskStep =
    case taskStep of
        ProjectTask.Build _ ->
            "Build"

        ProjectTask.Run _ ->
            "Run"

        ProjectTask.Clone _ ->
            "Clone"

        ProjectTask.Compose _ ->
            "Compose"

        ProjectTask.Push _ ->
            "Push"
