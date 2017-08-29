import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { APIService } from "./api.service";

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'app';

  constructor(
    private router: Router,
    private apiService: APIService
  ) {
  }

}
