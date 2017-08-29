import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import {NgbModule} from '@ng-bootstrap/ng-bootstrap';

import { AppComponent } from './app.component';
import { RouterModule } from '@angular/router';

import { IndexComponent } from './index/index.component';
import { SignInComponent } from "./authentication/sign-in.component";
import { APIService } from "./api.service";
import { AuthService } from "./authentication/authentication.service";
import { HttpModule } from "@angular/http";
import { ManageProjectsComponent } from "./projects/manage-projects.component";
import { ProjectRepository } from "./projects/project.repository";


@NgModule({
  declarations: [
    AppComponent,
    IndexComponent,
    SignInComponent,
    ManageProjectsComponent
  ],
  imports: [
    BrowserModule,
    NgbModule.forRoot(),
    FormsModule,
    ReactiveFormsModule,
    HttpModule,
    RouterModule.forRoot([
      { path: '', component: IndexComponent},
      { path: 'sign-in', component: SignInComponent},
      { path: 'manage-projects', component: ManageProjectsComponent}
    ], { useHash: true })
  ],
  providers: [
    AuthService,
    APIService,
    ProjectRepository,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
