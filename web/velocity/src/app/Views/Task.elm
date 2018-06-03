module Views.Task exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Data.Task as ProjectTask exposing (BuildStep, RunStep, CloneStep, ComposeStep, PushStep, Step(..), Parameter(..))


viewComposeStep : ComposeStep -> Bool -> Html msg
viewComposeStep step toggled =
    let
        title =
            "Compose" ++ step.description
    in
        viewStepCollapse (Compose step) title toggled <|
            []


viewPushStep : PushStep -> Bool -> Html msg
viewPushStep step toggled =
    let
        title =
            "Push" ++ step.description
    in
        viewStepCollapse (Push step) title toggled <|
            []


viewCloneStep : CloneStep -> Bool -> Html msg
viewCloneStep step toggled =
    let
        title =
            "Clone" ++ step.description
    in
        viewStepCollapse (Clone step) title toggled <|
            []


viewBuildStep : BuildStep -> Bool -> Html msg
viewBuildStep step toggled =
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
        viewStepCollapse (ProjectTask.Build step) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ rightDl ]
                ]
            ]


viewRunStep : Int -> RunStep -> Bool -> Html msg
viewRunStep i runStep toggled =
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
            toString i ++ ". " ++ runStep.description
    in
        viewStepCollapse (Run runStep) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ envTable ]
                ]
            ]


viewStepCollapse : Step -> String -> Bool -> List (Html msg) -> Html msg
viewStepCollapse step title toggled contents =
    let
        caretClassList =
            [ ( "fa-caret-square-o-down", toggled )
            , ( "fa-caret-square-o-up", not toggled )
            ]
    in
        div [ class "card" ]
            [ div [ class "card-header collapse-header d-flex justify-content-between align-items-center" ]
                [ h5 [ class "mb-0" ] [ text title ]
                , button
                    [ type_ "button"
                    , class "btn"
                    ]
                    [ i [ class "fa", classList caretClassList ] []
                    ]
                ]
            , div
                [ class "collapse"
                , classList [ ( "show", toggled ) ]
                ]
                [ div [ class "card-body" ] contents
                ]
            ]
