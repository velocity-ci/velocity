module Page.Project.Commit.Task exposing (..)

import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStream, BuildStreamOutput)
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Data.PaginatedList as PaginatedList exposing (Paginated(..))
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, on, onSubmit)
import Http
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (validClasses)
import Request.Commit
import Request.Build
import Task exposing (Task)
import Util exposing ((=>))
import Validate exposing (..)
import Views.Form as Form
import Views.Page as Page
import Json.Decode as Decode
import Html.Events.Extra exposing (targetSelectedIndex)
import Route
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Views.Task exposing (viewStepList)
import Views.Build exposing (viewBuildContainer)


-- MODEL --


type alias Model =
    { task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : List Field
    , errors : List Error
    , build : Maybe BuildType
    }


type Field
    = Input InputFormField
    | Choice ChoiceFormField


type StepType
    = LoadedBuildStep BuildStep


type BuildType
    = LoadedBuild Build (List BuildStep) (List BuildStreamOutput)


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


loadFirstBuild maybeAuthToken builds =
    case List.head builds of
        Just b ->
            Request.Build.steps b.id maybeAuthToken
                |> Http.toTask
                |> Task.andThen
                    (\(Paginated steps) ->
                        steps.results
                            |> List.map (.id >> (Request.Build.streams maybeAuthToken) >> Http.toTask)
                            |> Task.sequence
                            |> Task.andThen
                                (\paginatedStreams ->
                                    paginatedStreams
                                        |> List.map (\(Paginated { results }) -> results)
                                        |> List.foldr (++) []
                                        |> List.map (.id >> (Request.Build.streamOutput maybeAuthToken) >> Http.toTask)
                                        |> Task.sequence
                                        |> Task.andThen
                                            (\paginatedStreamOutputList ->
                                                paginatedStreamOutputList
                                                    |> List.map (\(Paginated { results }) -> results)
                                                    |> List.foldr (++) []
                                                    |> LoadedBuild b steps.results
                                                    |> Just
                                                    |> Task.succeed
                                            )
                                )
                    )

        Nothing ->
            Task.succeed Nothing


init : Session msg -> Project.Id -> Commit.Hash -> ProjectTask.Name -> List Build -> Task PageLoadError Model
init session id hash name builds =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadTask =
            maybeAuthToken
                |> Request.Commit.task id hash name
                |> Http.toTask

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."

        initialModel task build =
            let
                toggledStep =
                    Nothing

                form =
                    List.map newField task.parameters

                errors =
                    List.concatMap validator form
            in
                { task = task
                , toggledStep = toggledStep
                , form = form
                , errors = errors
                , build = build
                }

        loadBuild =
            loadFirstBuild maybeAuthToken builds
    in
        Task.map2 initialModel loadTask loadBuild
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

        buildContainer =
            case model.build of
                Just (LoadedBuild build steps output) ->
                    viewBuildContainer build steps output

                Nothing ->
                    buildForm
    in
        div [ class "row" ]
            [ div [ class "col-sm-12 col-md-12 col-lg-12 default-margin-bottom" ]
                [ buildContainer ]
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
        [ h4 []
            [ a []
                [ text taskName ]
            ]
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


breadcrumb : Project -> Commit -> ProjectTask.Task -> List ( Route.Route, String )
breadcrumb project commit task =
    [ ( CommitRoute.Task task.name |> ProjectRoute.Commit commit.hash |> Route.Project project.id
      , ProjectTask.nameToString task.name
      )
    ]



-- UPDATE --


type Msg
    = ToggleStep (Maybe Step)
    | OnInput InputFormField String
    | OnChange ChoiceFormField (Maybe Int)
    | SubmitForm
    | BuildCreated (Result Http.Error Build)


update : Project -> Commit -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project commit session msg model =
    let
        projectId =
            project.id

        commitHash =
            commit.hash

        taskName =
            model.task.name
    in
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
                            |> Request.Commit.createBuild projectId commitHash taskName params
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
                                choice.value
                                    |> Maybe.map (\value -> stringParm { value = value, field = choice.field })

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
