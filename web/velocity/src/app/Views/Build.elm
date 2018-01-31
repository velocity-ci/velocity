module Views.Build exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStreamOutput)
import Ansi.Log


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
            i [ class "fa fa-cog fa-spin" ] []

        BuildStep.Running ->
            i [ class "fa fa-cog fa-spin" ] []

        BuildStep.Success ->
            i [ class "fa fa-check" ] []

        BuildStep.Failed ->
            i [ class "fa fa-times" ] []
