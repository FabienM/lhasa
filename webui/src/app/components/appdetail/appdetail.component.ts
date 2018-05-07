import { Component, OnInit } from '@angular/core';

import { ActivatedRoute } from '@angular/router';
import { ApplicationBean, DeploymentBean, EnvironmentBean } from '../../models/commons/applications-bean';
import { Store } from '@ngrx/store';
import { ApplicationsStoreService } from '../../stores/applications-store.service';
import { EnvironmentsStoreService } from '../../stores/environments-store.service';
import { element } from 'protractor';
import { ISubscription } from 'rxjs/Subscription';
import { AutoUnsubscribe } from '../../shared/decorator/autoUnsubscribe';


@Component({
  selector: 'app-appdetail',
  templateUrl: './appdetail.component.html',
  styleUrls: ['./appdetail.component.css'],

})
@AutoUnsubscribe()
export class AppdetailComponent implements OnInit {

  /**
   * internal streams and store
   */
  protected applicationStream: Store<ApplicationBean>;
  protected applicationSubscription: ISubscription;
  protected deploymentStream: Store<DeploymentBean[]>;
  protected deploymentSubscription: ISubscription;
  public application: ApplicationBean;
  protected deployments: DeploymentBean[];

  constructor(
    private applicationsStoreService: ApplicationsStoreService,
    private route: ActivatedRoute) {
    /**
     * subscribe
     */
    this.applicationStream = this.applicationsStoreService.active();
    this.deploymentStream = this.applicationsStoreService.deployments();
  }

  ngOnInit(): void {
    this.applicationSubscription = this.applicationStream.subscribe(
      (app: ApplicationBean) => {
        this.application = app;
      },
      error => {
        console.error(error);
      },
      () => {
      }
    );

    this.deploymentSubscription = this.deploymentStream.subscribe(
      (app: DeploymentBean[]) => {
        this.deployments = app;
      },
      error => {
        console.error(error);
      },
      () => {
      }
    );
  }
}
