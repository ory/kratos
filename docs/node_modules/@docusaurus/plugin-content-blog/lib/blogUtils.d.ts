/**
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
import { Feed } from 'feed';
import { PluginOptions, BlogPost } from './types';
import { LoadContext } from '@docusaurus/types';
export declare function truncate(fileString: string, truncateMarker: RegExp): string;
export declare function generateBlogFeed(context: LoadContext, options: PluginOptions): Promise<Feed | null>;
export declare function generateBlogPosts(blogDir: string, { siteConfig, siteDir }: LoadContext, options: PluginOptions): Promise<BlogPost[]>;
export declare function linkify(fileContent: string, siteDir: string, blogPath: string, blogPosts: BlogPost[]): string;
//# sourceMappingURL=blogUtils.d.ts.map