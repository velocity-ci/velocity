module Page.Project.Task exposing (..)

import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build
import Data.Task as ProjectTask exposing (BuildStep, RunStep, CloneStep, Step(..), Parameter(..))
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Http
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (validClasses)
import Page.Project.Commit as Commit
import Page.Project.Route as ProjectRoute
import Request.Commit
import Route exposing (Route)
import Task exposing (Task)
import Util exposing ((=>))
import Validate exposing (..)
import Views.Form as Form
import Views.Page as Page
import Json.Decode as Decode
import Html.Events.Extra exposing (targetSelectedIndex)


-- MODEL --


type alias Model =
    { commit : Commit
    , task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : List Field
    , errors : List Error
    }


type Field
    = Input InputFormField
    | Choice ChoiceFormField


type alias InputFormField =
    { value : String
    , dirty : Bool
    , field : String
    }


type alias ChoiceFormField =
    { value : Maybe String
    , dirty : Bool
    , field : String
    , options : List String
    }


init : Session -> Project.Id -> Commit.Hash -> ProjectTask.Name -> Task PageLoadError Model
init session id hash name =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommit =
            maybeAuthToken
                |> Request.Commit.get id hash
                |> Http.toTask

        loadTask =
            maybeAuthToken
                |> Request.Commit.task id hash name
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."

        initialModel commit task =
            let
                toggledStep =
                    Nothing

                form =
                    List.map newField task.parameters

                errors =
                    List.concatMap validator form
            in
                { commit = commit
                , task = task
                , toggledStep = toggledStep
                , form = form
                , errors = errors
                }
    in
        Task.map2 initialModel loadCommit loadTask
            |> Task.mapError handleLoadError


newField : Parameter -> Field
newField parameter =
    case parameter of
        StringParam param ->
            let
                value =
                    Maybe.withDefault "" param.default

                dirty =
                    String.length value > 0
            in
                InputFormField value dirty param.name
                    |> Input

        ChoiceParam param ->
            let
                options =
                    param.default
                        :: (List.map Just param.options)
                        |> List.filterMap identity

                value =
                    case param.default of
                        Nothing ->
                            List.head options

                        default ->
                            default
            in
                ChoiceFormField value True param.name options
                    |> Choice



-- VIEW --


view : Model -> Html Msg
view model =
    let
        task =
            model.task

        stepList =
            viewStepList task.steps model.toggledStep

        buildForm =
            div [ class "card" ]
                [ div [ class "card-body" ] <|
                    viewBuildForm (ProjectTask.nameToString task.name) model.form model.errors
                ]
    in
        div [ class "row" ]
            [ div [ class "col-sm-12 col-md-12 col-lg-12 default-margin-bottom" ] [ buildForm ]
            , div [ class "col-sm-12 col-md-12 col-lg-12" ] stepList
            ]


viewBuildForm : String -> List Field -> List Error -> List (Html Msg)
viewBuildForm taskName fields errors =
    let
        fieldInput f =
            case f of
                Choice field ->
                    let
                        value =
                            Maybe.withDefault "" field.value

                        option o =
                            Html.option
                                [ selected (o == value) ]
                                [ text o ]
                    in
                        Form.select
                            { name = field.field
                            , label = field.field
                            , help = Nothing
                            , errors = []
                            }
                            [ attribute "required" ""
                            , classList (validClasses errors field)
                            , on "change" <| Decode.map (OnChange field) targetSelectedIndex
                            ]
                            (List.map option field.options)

                Input field ->
                    Form.input
                        { name = field.field
                        , label = field.field
                        , help = Nothing
                        , errors = []
                        }
                        [ attribute "required" ""
                        , value field.value
                        , onInput (OnInput field)
                        , classList (validClasses errors field)
                        ]
                        []
    in
        [ h4 [] [ text taskName ]
        , Html.form [ attribute "novalidate" "", onSubmit SubmitForm ] <|
            List.map fieldInput fields
                ++ [ button
                        [ class "btn btn-primary"
                        , type_ "submit"
                        , disabled <| not (List.isEmpty errors)
                        ]
                        [ text "Start task" ]
                   ]
        ]


viewStepList : List Step -> Maybe Step -> List (Html Msg)
viewStepList steps toggledStep =
    let
        stepView i step =
            let
                stepNum =
                    i + 1

                runStep =
                    viewRunStep stepNum

                buildStep =
                    viewBuildStep stepNum

                cloneStep =
                    viewCloneStep stepNum
            in
                case ( step, toggledStep ) of
                    ( Run run, Just (Run toggled) ) ->
                        runStep run (run == toggled)

                    ( Build build, Just (Build toggled) ) ->
                        buildStep build (build == toggled)

                    ( Clone clone, Just (Clone toggled) ) ->
                        cloneStep clone (clone == toggled)

                    ( Run run, _ ) ->
                        runStep run False

                    ( Build build, _ ) ->
                        buildStep build False

                    ( Clone clone, _ ) ->
                        cloneStep clone False
    in
        List.indexedMap stepView steps


viewCloneStep : Int -> CloneStep -> Bool -> Html Msg
viewCloneStep i cloneStep toggled =
    let
        title =
            toString i ++ ". Clone" ++ cloneStep.description
    in
        viewStepCollapse (Clone cloneStep) title toggled <|
            []


viewBuildStep : Int -> BuildStep -> Bool -> Html Msg
viewBuildStep i buildStep toggled =
    let
        tagList =
            List.map (\t -> li [] [ text t ]) buildStep.tags
                |> ul []

        rightDl =
            dl []
                [ dt [] [ text "Tags" ]
                , dd [] [ tagList ]
                ]

        leftDl =
            dl []
                [ dt [] [ text "Context" ]
                , dd [] [ text buildStep.context ]
                , dt [] [ text "Dockerfile" ]
                , dd [] [ text buildStep.dockerfile ]
                ]

        title =
            toString i ++ ". " ++ buildStep.description
    in
        viewStepCollapse (Build buildStep) title toggled <|
            [ div [ class "row" ]
                [ div [ class "col-md-6" ] [ leftDl ]
                , div [ class "col-md-6" ] [ rightDl ]
                ]
            ]


viewRunStep : Int -> RunStep -> Bool -> Html Msg
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


viewStepCollapse : Step -> String -> Bool -> List (Html Msg) -> Html Msg
viewStepCollapse step title toggled contents =
    let
        msg =
            if toggled then
                ToggleStep Nothing
            else
                ToggleStep (Just step)

        caretClassList =
            [ ( "fa-caret-square-o-down", toggled )
            , ( "fa-caret-square-o-up", not toggled )
            ]
    in
        div [ class "card" ]
            [ div [ class "card-header collapse-header d-flex justify-content-between align-items-center", onClick msg ]
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


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route, String )
breadcrumb project commit task =
    List.concat
        [ Commit.breadcrumb project commit
        , [ ( Route.Project project.id (ProjectRoute.Task commit.hash task.name), ProjectTask.nameToString task.name ) ]
        ]



-- UPDATE --


type Msg
    = ToggleStep (Maybe Step)
    | OnInput InputFormField String
    | OnChange ChoiceFormField (Maybe Int)
    | SubmitForm
    | BuildCreated (Result Http.Error Data.Build.Build)


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    case msg of
        ToggleStep maybeStep ->
            { model | toggledStep = maybeStep }
                => Cmd.none

        OnInput field value ->
            let
                updateField fieldType =
                    case fieldType of
                        Input f ->
                            if f == field then
                                Input
                                    { field
                                        | value = value
                                        , dirty = True
                                    }
                            else
                                fieldType

                        _ ->
                            fieldType

                form =
                    List.map updateField model.form

                errors =
                    List.concatMap validator form
            in
                { model
                    | form = form
                    , errors = errors
                }
                    => Cmd.none

        OnChange field maybeIndex ->
            let
                updateField fieldType =
                    case ( fieldType, maybeIndex ) of
                        ( Choice f, Just index ) ->
                            if f == field then
                                let
                                    value =
                                        f.options
                                            |> List.indexedMap (,)
                                            |> List.filter (\i -> Tuple.first i == index)
                                            |> List.head
                                            |> Maybe.map Tuple.second
                                in
                                    Choice
                                        { field
                                            | value = value
                                            , dirty = True
                                        }
                            else
                                fieldType

                        _ ->
                            fieldType

                form =
                    List.map updateField model.form

                errors =
                    List.concatMap validator form
            in
                { model
                    | form = form
                    , errors = errors
                }
                    => Cmd.none

        SubmitForm ->
            let
                stringParm { value, field } =
                    field => value

                cmdFromAuth authToken =
                    authToken
                        |> Request.Commit.createBuild project.id model.commit.hash model.task.name params
                        |> Http.send BuildCreated

                cmd =
                    session
                        |> Session.attempt "create build" cmdFromAuth
                        |> Tuple.second

                mapFieldToParam field =
                    case field of
                        Input input ->
                            Just (stringParm input)

                        Choice choice ->
                            case choice.value of
                                Just value ->
                                    Just (stringParm { value = value, field = choice.field })

                                Nothing ->
                                    Nothing

                params =
                    List.filterMap mapFieldToParam model.form
            in
                model => cmd

        BuildCreated _ ->
            model => Cmd.none



-- VALIDATION --


type alias Error =
    ( String, String )


validator : Validator Error Field
validator =
    [ \f ->
        let
            notBlank { field, value } =
                ifBlank (field => "Field cannot be blank") value
        in
            case f of
                Input fieldType ->
                    notBlank fieldType

                Choice fieldType ->
                    (\{ field, value } ->
                        value
                            |> Maybe.withDefault ""
                            |> ifBlank (field => "Field cannot be blank")
                    )
                        fieldType
    ]
        |> Validate.all
