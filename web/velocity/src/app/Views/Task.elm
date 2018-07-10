module Views.Task exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Data.Task as ProjectTask exposing (BuildStep, RunStep, CloneStep, ComposeStep, PushStep, Step(..), Parameter(..))


viewComposeStep : ComposeStep -> Html msg
viewComposeStep step =
    let
        title =
            "Compose" ++ step.description
    in
        div [] []


viewPushStep : PushStep -> Html msg
viewPushStep step =
    let
        title =
            "Push" ++ step.description
    in
        div [] []


viewCloneStep : CloneStep -> Html msg
viewCloneStep step =
    let
        title =
            "Clone" ++ step.description
    in
        div [] []


viewBuildStep : BuildStep -> Html msg
viewBuildStep step =
    let
        tagList =
            List.map (\t -> li [] [ text t ]) step.tags
                |> ul []

        rightDl =
            dl []
                [ dt [] [ text "Tags" ]
                , dd [] [ tagList ]
                ]

        leftDl =
            dl []
                [ dt [] [ text "Context" ]
                , dd [] [ text step.context ]
                , dt [] [ text "Dockerfile" ]
                , dd [] [ text step.dockerfile ]
                ]

        title =
            "Build" ++ step.description
    in
        div [ class "row" ]
            [ div [ class "col-md-6" ] [ leftDl ]
            , div [ class "col-md-6" ] [ rightDl ]
            ]


viewRunStep : RunStep -> Html msg
viewRunStep runStep =
    let
        command =
            String.join " " runStep.command

        envTable =
            table [ class "table" ]
                [ tbody []
                    (List.map
                        (\( k, v ) ->
                            tr []
                                [ th [] [ text k ]
                                , td [] [ text v ]
                                ]
                        )
                        runStep.environment
                    )
                ]

        ignoreExitCode =
            runStep.ignoreExitCode
                |> toString
                |> String.toLower

        leftDl =
            dl []
                [ dt [] [ text "Ignore exit code" ]
                , dd [] [ text ignoreExitCode ]
                , dt [] [ text "Image" ]
                , dd [] [ text runStep.image ]
                , dt [] [ text "Mount point" ]
                , dd [] [ text runStep.mountPoint ]
                , dt [] [ text "Working dir" ]
                , dd [] [ text runStep.workingDir ]
                , dt [] [ text "Command" ]
                , dd [] [ text command ]
                ]

        title =
            runStep.description
    in
        div [ class "row" ]
            [ div [ class "col-md-6" ] [ leftDl ]
            , div [ class "col-md-6" ] [ envTable ]
            ]


viewStepContents : Step -> Html msg
viewStepContents step =
    case step of
        Compose composeStep ->
            viewComposeStep composeStep

        Push pushStep ->
            viewPushStep pushStep

        Clone cloneStep ->
            viewCloneStep cloneStep

        Build buildStep ->
            viewBuildStep buildStep

        Run runStep ->
            viewRunStep runStep
