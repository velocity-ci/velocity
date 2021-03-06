-- Do not manually edit this file, it was auto-generated by dillonkearns/elm-graphql
-- https://github.com/dillonkearns/elm-graphql


module Api.Compiled.Query exposing (BranchRequiredArguments, CommitsOptionalArguments, CommitsRequiredArguments, EventsOptionalArguments, ProjectRequiredArguments, ProjectsOptionalArguments, branch, commits, events, listKnownHosts, project, projects)

import Api.Compiled.InputObject
import Api.Compiled.Interface
import Api.Compiled.Object
import Api.Compiled.Scalar
import Api.Compiled.ScalarCodecs
import Api.Compiled.Union
import Graphql.Internal.Builder.Argument as Argument exposing (Argument)
import Graphql.Internal.Builder.Object as Object
import Graphql.Internal.Encode as Encode exposing (Value)
import Graphql.Operation exposing (RootMutation, RootQuery, RootSubscription)
import Graphql.OptionalArgument exposing (OptionalArgument(..))
import Graphql.SelectionSet exposing (SelectionSet)
import Json.Decode as Decode exposing (Decoder)


type alias BranchRequiredArguments =
    { branch : String
    , projectSlug : String
    }


{-| Get branch
-}
branch : BranchRequiredArguments -> SelectionSet decodesTo Api.Compiled.Object.Branch -> SelectionSet decodesTo RootQuery
branch requiredArgs object_ =
    Object.selectionForCompositeField "branch" [ Argument.required "branch" requiredArgs.branch Encode.string, Argument.required "projectSlug" requiredArgs.projectSlug Encode.string ] object_ identity


type alias CommitsOptionalArguments =
    { after : OptionalArgument String
    , before : OptionalArgument String
    , first : OptionalArgument Int
    , last : OptionalArgument Int
    }


type alias CommitsRequiredArguments =
    { branch : String
    , projectSlug : String
    }


{-| List commits
-}
commits : (CommitsOptionalArguments -> CommitsOptionalArguments) -> CommitsRequiredArguments -> SelectionSet decodesTo Api.Compiled.Object.CommitConnection -> SelectionSet (Maybe decodesTo) RootQuery
commits fillInOptionals requiredArgs object_ =
    let
        filledInOptionals =
            fillInOptionals { after = Absent, before = Absent, first = Absent, last = Absent }

        optionalArgs =
            [ Argument.optional "after" filledInOptionals.after Encode.string, Argument.optional "before" filledInOptionals.before Encode.string, Argument.optional "first" filledInOptionals.first Encode.int, Argument.optional "last" filledInOptionals.last Encode.int ]
                |> List.filterMap identity
    in
    Object.selectionForCompositeField "commits" (optionalArgs ++ [ Argument.required "branch" requiredArgs.branch Encode.string, Argument.required "projectSlug" requiredArgs.projectSlug Encode.string ]) object_ (identity >> Decode.nullable)


type alias EventsOptionalArguments =
    { after : OptionalArgument String
    , before : OptionalArgument String
    , first : OptionalArgument Int
    , last : OptionalArgument Int
    }


{-| List events
-}
events : (EventsOptionalArguments -> EventsOptionalArguments) -> SelectionSet decodesTo Api.Compiled.Object.EventConnection -> SelectionSet (Maybe decodesTo) RootQuery
events fillInOptionals object_ =
    let
        filledInOptionals =
            fillInOptionals { after = Absent, before = Absent, first = Absent, last = Absent }

        optionalArgs =
            [ Argument.optional "after" filledInOptionals.after Encode.string, Argument.optional "before" filledInOptionals.before Encode.string, Argument.optional "first" filledInOptionals.first Encode.int, Argument.optional "last" filledInOptionals.last Encode.int ]
                |> List.filterMap identity
    in
    Object.selectionForCompositeField "events" optionalArgs object_ (identity >> Decode.nullable)


{-| Get all known hosts
-}
listKnownHosts : SelectionSet decodesTo Api.Compiled.Object.KnownHost -> SelectionSet (List decodesTo) RootQuery
listKnownHosts object_ =
    Object.selectionForCompositeField "listKnownHosts" [] object_ (identity >> Decode.list)


type alias ProjectRequiredArguments =
    { slug : String }


{-| Get project
-}
project : ProjectRequiredArguments -> SelectionSet decodesTo Api.Compiled.Object.Project -> SelectionSet decodesTo RootQuery
project requiredArgs object_ =
    Object.selectionForCompositeField "project" [ Argument.required "slug" requiredArgs.slug Encode.string ] object_ identity


type alias ProjectsOptionalArguments =
    { after : OptionalArgument String
    , before : OptionalArgument String
    , first : OptionalArgument Int
    , last : OptionalArgument Int
    }


{-| List projects
-}
projects : (ProjectsOptionalArguments -> ProjectsOptionalArguments) -> SelectionSet decodesTo Api.Compiled.Object.ProjectConnection -> SelectionSet (Maybe decodesTo) RootQuery
projects fillInOptionals object_ =
    let
        filledInOptionals =
            fillInOptionals { after = Absent, before = Absent, first = Absent, last = Absent }

        optionalArgs =
            [ Argument.optional "after" filledInOptionals.after Encode.string, Argument.optional "before" filledInOptionals.before Encode.string, Argument.optional "first" filledInOptionals.first Encode.int, Argument.optional "last" filledInOptionals.last Encode.int ]
                |> List.filterMap identity
    in
    Object.selectionForCompositeField "projects" optionalArgs object_ (identity >> Decode.nullable)
