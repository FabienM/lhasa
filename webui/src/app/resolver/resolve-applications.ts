import { Observable } from 'rxjs/Observable';
import { Resolve, ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Injectable } from '@angular/core';
import { ApplicationBean, ApplicationPagesBean } from '../models/commons/applications-bean';
import { LoadDomainsAction, ApplicationsStoreService, LoadApplicationsAction } from '../stores/applications-store.service';
import { DataApplicationService } from '../services/data-application-version.service';
import { ContentListResponse, PageMetaData } from '../models/commons/entity-bean';
import { DataDeploymentService } from '../services/data-deployment.service';
import { DataDomainService } from '../services/data-domain.service';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';
import { Subject } from 'rxjs/Subject';

@Injectable()
export class ApplicationsResolver implements Resolve<ApplicationPagesBean> {
    constructor(
        private applicationsStoreService: ApplicationsStoreService,
        private applicationsService: DataApplicationService
    ) { }

    resolve(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<any> | Promise<any> | any {
        let domain = '';
        // If any domain on url use it
        if (route.queryParams.domain) {
            domain = '/' + route.queryParams.domain;
        }
        // Select application
        return this.selectApplications({
            number: route.queryParams.page || 0,
            size: 100
        }, domain, new BehaviorSubject<any>('select all applications'));
    }

    /**
     * dispatch load domains
     * @param event
     */
    public selectApplications(metadata: PageMetaData, domain: string, subject: Subject<any>): Subject<any> {
        // load all domains
        const meta: {
            [key: string]: any | any[];
        } = {
                size: metadata.size,
                page: metadata.number
            };
        this.applicationsService.GetAllFromContent(domain, meta).subscribe(
            (data: ContentListResponse<ApplicationBean>) => {
                this.applicationsStoreService.dispatch(
                    new LoadApplicationsAction({
                        applications: data.content,
                        metadata: data.pageMetadata,
                    }, subject)
                );
            }
        );
        return subject;
    }
}
