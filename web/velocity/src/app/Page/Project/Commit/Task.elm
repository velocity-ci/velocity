module Page.Project.Commit.Task exposing (..)

import Ansi.Log
import Data.Commit as Commit exposing (Commit)
import Data.Project as Project exposing (Project)
import Data.Session as Session exposing (Session)
import Data.Build as Build exposing (Build)
import Data.BuildStep as BuildStep exposing (BuildStep)
import Data.BuildStream as BuildStream exposing (BuildStream, BuildStreamOutput)
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Data.PaginatedList as PaginatedList exposing (Paginated(..))
import Data.AuthToken as AuthToken exposing (AuthToken)
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
import Views.Helpers exposing (onClickPage)
import Navigation
import Dict exposing (Dict)
import Json.Encode as Encode


-- MODEL --


type alias Model =
    { task : ProjectTask.Task
    , toggledStep : Maybe Step
    , form : List Field
    , errors : List Error
    , selectedTab : Tab
    , frame : Frame
    }


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


type Field
    = Input InputFormField
    | Choice ChoiceFormField


type StepType
    = LoadedBuildStep BuildStep


type alias FromBuild =
    Build


type alias ToBuild =
    Build


type BuildType
    = LoadedBuild Build (List BuildStep) (List BuildStreamOutput) Ansi.Log.Model
    | LoadingBuild (Maybe FromBuild) ToBuild


type Tab
    = NewFormTab
    | BuildTab Build


type Frame
    = BuildFrame BuildType
    | NewFormFrame


loadBuild :
    Maybe AuthToken
    -> Build
    -> Task Http.Error (Maybe BuildType)
loadBuild maybeAuthToken build =
    Request.Build.steps build.id maybeAuthToken
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
                                            |> (\outputStreams ->
                                                    buildOutput outputStreams
                                                        |> LoadedBuild build steps.results outputStreams
                                               )
                                            |> Just
                                            |> Task.succeed
                                    )
                        )
            )


loadFirstBuild :
    Maybe AuthToken
    -> List Build
    -> Task Http.Error (Maybe BuildType)
loadFirstBuild maybeAuthToken builds =
    List.head builds
        |> Maybe.map (loadBuild maybeAuthToken)
        |> Maybe.withDefault (Task.succeed Nothing)


buildOutput : List BuildStreamOutput -> Ansi.Log.Model
buildOutput buildOutput =
    List.foldl Ansi.Log.update
        (Ansi.Log.init Ansi.Log.Cooked)
        (List.map .output buildOutput)


stringToTab : Maybe String -> List Build -> Tab
stringToTab maybeSelectedTab builds =
    case maybeSelectedTab of
        Just "new" ->
            NewFormTab

        Just buildId ->
            builds
                |> List.filter (\b -> (Build.idToString b.id) == buildId)
                |> List.head
                |> Maybe.map BuildTab
                |> Maybe.withDefault NewFormTab

        Nothing ->
            NewFormTab


init : Session msg -> Project.Id -> Commit.Hash -> ProjectTask.Task -> Maybe String -> List Build -> Task PageLoadError Model
init session id hash task maybeSelectedTab builds =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."

        selectedTab =
            stringToTab maybeSelectedTab builds

        initialModel frame =
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
                , selectedTab = selectedTab
                , frame = Maybe.withDefault NewFormFrame frame
                }
    in
        case selectedTab of
            NewFormTab ->
                Task.succeed (initialModel (Just NewFormFrame))

            BuildTab b ->
                loadBuild maybeAuthToken b
                    |> Task.andThen (Maybe.map BuildFrame >> Task.succeed)
                    |> Task.map initialModel
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



-- CHANNELS --


streamChannelName : BuildStream -> String
streamChannelName stream =
    "stream:" ++ (BuildStream.idToString stream.id)


buildEvents : Build -> Dict String (List ( String, Encode.Value -> Msg ))
buildEvents build =
    let
        streams =
            List.map .streams build.steps
                |> List.foldr (++) []

        foldStreamEvents stream dict =
            let
                channelName =
                    streamChannelName stream

                events =
                    [ ( "streamLine:new", AddStreamOutput stream ) ]
            in
                Dict.insert channelName events dict
    in
        List.foldl foldStreamEvents Dict.empty streams


events : Model -> Dict String (List ( String, Encode.Value -> Msg ))
events model =
    case model.frame of
        BuildFrame (LoadedBuild build _ _ _) ->
            buildEvents build

        _ ->
            Dict.empty


leaveChannels : Model -> Maybe String -> List String
leaveChannels model maybeBuildId =
    let
        channels id b =
            if id == Build.idToString b.id then
                []
            else
                Dict.keys (buildEvents b)
    in
        case ( maybeBuildId, model.frame ) of
            ( Just buildId, BuildFrame (LoadedBuild b _ _ _) ) ->
                channels buildId b

            ( Just buildId, BuildFrame (LoadingBuild (Just b) _) ) ->
                channels buildId b

            ( _, BuildFrame (LoadedBuild b _ _ _) ) ->
                Dict.keys (buildEvents b)

            ( _, BuildFrame (LoadingBuild (Just b) _) ) ->
                Dict.keys (buildEvents b)

            _ ->
                []



-- VIEW --


view : Project -> Commit -> Model -> List Build -> Html Msg
view project commit model builds =
    let
        task =
            model.task

        stepList =
            viewStepList task.steps model.toggledStep
    in
        div [ class "row" ]
            [ div [ class "col-sm-12 col-md-12 col-lg-12 default-margin-bottom" ]
                [ h4 [] [ text (ProjectTask.nameToString task.name) ]
                , viewTabs project commit task builds model.selectedTab
                , viewTabFrame model
                ]
            ]


viewTabs : Project -> Commit -> ProjectTask.Task -> List Build -> Tab -> Html Msg
viewTabs project commit task builds selectedTab =
    let
        buildTab t =
            let
                tabText =
                    case t of
                        NewFormTab ->
                            "+"

                        BuildTab b ->
                            Build.idToString b.id
                                |> String.slice 0 5

                tabQueryParam =
                    case t of
                        NewFormTab ->
                            "new"

                        BuildTab b ->
                            Build.idToString b.id

                route =
                    CommitRoute.Task task.name (Just tabQueryParam)
                        |> ProjectRoute.Commit commit.hash
                        |> Route.Project project.id

                compare a b =
                    case ( a, b ) of
                        ( BuildTab c, BuildTab d ) ->
                            Build.idToString c.id == Build.idToString d.id

                        ( NewFormTab, NewFormTab ) ->
                            True

                        _ ->
                            False

                tabClassList =
                    [ ( "nav-link", True )
                    , ( "active", compare t selectedTab )
                    ]
            in
                li [ class "nav-item" ]
                    [ a
                        [ classList tabClassList
                        , Route.href route
                        , onClickPage (SelectTab selectedTab) route
                        ]
                        [ text tabText ]
                    ]
    in
        List.append (List.map (BuildTab >> buildTab) builds) [ buildTab NewFormTab ]
            |> ul [ class "nav nav-tabs nav-fill" ]


viewTabFrame : Model -> Html Msg
viewTabFrame model =
    let
        buildForm =
            div [] <|
                viewBuildForm (ProjectTask.nameToString model.task.name) model.form model.errors
    in
        case model.frame of
            NewFormFrame ->
                buildForm

            BuildFrame (LoadedBuild _ _ _ o) ->
                Ansi.Log.view o

            BuildFrame (LoadingBuild _ _) ->
                text "Loading"


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
        [ Html.form [ attribute "novalidate" "", onSubmit SubmitForm ] <|
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
    [ ( CommitRoute.Task task.name Nothing |> ProjectRoute.Commit commit.hash |> Route.Project project.id
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
    | SelectTab Tab String
    | LoadBuild Build
    | BuildLoaded (Result Http.Error (Maybe BuildType))
    | AddStreamOutput BuildStream Encode.Value


type ExternalMsg
    = NoOp
    | AddBuild Build


update : Project -> Commit -> Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update project commit session msg model =
    let
        projectId =
            project.id

        commitHash =
            commit.hash

        taskName =
            model.task.name

        maybeAuthToken =
            Maybe.map .token session.user
    in
        case msg of
            ToggleStep maybeStep ->
                { model | toggledStep = maybeStep }
                    => Cmd.none
                    => NoOp

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
                        => NoOp

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
                                                |> List.filter (\( i, _ ) -> i == index)
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
                        => NoOp

            SubmitForm ->
                let
                    stringParam { value, field } =
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
                                Just (stringParam input)

                            Choice choice ->
                                choice.value
                                    |> Maybe.map (\value -> stringParam { value = value, field = choice.field })

                    params =
                        List.filterMap mapFieldToParam model.form
                in
                    model
                        => cmd
                        => NoOp

            LoadBuild build ->
                let
                    cmd =
                        build
                            |> loadBuild maybeAuthToken
                            |> Task.attempt BuildLoaded
                in
                    model
                        => cmd
                        => NoOp

            BuildLoaded (Ok (Just loadedBuild)) ->
                model
                    => Cmd.none
                    => NoOp

            BuildLoaded _ ->
                model
                    => Cmd.none
                    => NoOp

            BuildCreated (Ok build) ->
                let
                    tab =
                        build.id
                            |> Build.idToString
                            |> Just

                    route =
                        CommitRoute.Task model.task.name tab
                            |> ProjectRoute.Commit commit.hash
                            |> Route.Project project.id
                in
                    model
                        => Navigation.newUrl (Route.routeToString route)
                        => AddBuild build

            BuildCreated (Err _) ->
                model
                    => Cmd.none
                    => NoOp

            SelectTab tab url ->
                let
                    frame =
                        case tab of
                            BuildTab toBuild ->
                                let
                                    fromBuild =
                                        case model.frame of
                                            BuildFrame (LoadedBuild b _ _ _) ->
                                                Just b

                                            _ ->
                                                Nothing
                                in
                                    BuildFrame (LoadingBuild fromBuild toBuild)

                            NewFormTab ->
                                NewFormFrame
                in
                    { model
                        | selectedTab = tab
                        , frame = frame
                    }
                        => Navigation.newUrl url
                        => NoOp

            AddStreamOutput _ outputJson ->
                let
                    maybeBuildStreamOutput =
                        outputJson
                            |> Decode.decodeValue BuildStream.outputDecoder
                            |> Result.toMaybe

                    outputStreams existing =
                        maybeBuildStreamOutput
                            |> Maybe.map (\b -> List.append existing [ b ])
                            |> Maybe.withDefault existing

                    output existingOutput =
                        maybeBuildStreamOutput
                            |> Maybe.map (\o -> Ansi.Log.update o.output existingOutput)
                            |> Maybe.withDefault existingOutput

                    frame =
                        case model.frame of
                            BuildFrame (LoadedBuild a b c d) ->
                                BuildFrame (LoadedBuild a b (outputStreams c) (buildOutput c))

                            _ ->
                                model.frame
                in
                    { model | frame = frame }
                        => Cmd.none
                        => NoOp



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
